package database_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"bimbingan-backend/internal/migration"

	_ "github.com/lib/pq"
)

func TestPostgreSQLIntegrationAndConstraints(t *testing.T) {
	testURL := os.Getenv("TEST_DATABASE_URL")
	if testURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping live PostgreSQL integration tests")
	}

	// Safety Guard: Require '_test' in database URL or ALLOW_TEST_DATABASE_RESET=true
	allowReset := os.Getenv("ALLOW_TEST_DATABASE_RESET") == "true"
	if !strings.Contains(strings.ToLower(testURL), "_test") && !allowReset {
		t.Fatal("SAFETY BLOCK: Destructive migration test rejected because TEST_DATABASE_URL does not contain '_test' and ALLOW_TEST_DATABASE_RESET is not true")
	}

	// 1. Run Migration Lifecycle: Up -> DownAll -> Up
	runner := migration.NewRunner(testURL)

	t.Log("running initial migration Up...")
	if err := runner.Up(); err != nil {
		t.Fatalf("migration Up failed: %v", err)
	}

	t.Log("testing migration DownAll...")
	if err := runner.DownAll(); err != nil {
		t.Fatalf("migration DownAll failed: %v", err)
	}

	t.Log("re-running migration Up...")
	if err := runner.Up(); err != nil {
		t.Fatalf("re-running migration Up failed: %v", err)
	}

	// Connect to database for constraint testing
	db, err := sql.Open("postgres", testURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Clean test table contents before negative constraint tests
	_, _ = db.ExecContext(ctx, "TRUNCATE TABLE annotation_snapshots, annotations, documents, guidance_sessions, refresh_sessions, users CASCADE;")

	t.Run("constraint: case-insensitive duplicate email is rejected", func(t *testing.T) {
		userID1 := "11111111-1111-1111-1111-111111111111"
		userID2 := "22222222-2222-2222-2222-222222222222"

		_, err := db.ExecContext(ctx, `INSERT INTO users (id, full_name, email, password_hash, role) VALUES ($1, $2, $3, $4, $5)`,
			userID1, "User One", "TEST.USER@EXAMPLE.COM", "hash1", "mahasiswa")
		if err != nil {
			t.Fatalf("failed to insert initial user: %v", err)
		}

		_, err = db.ExecContext(ctx, `INSERT INTO users (id, full_name, email, password_hash, role) VALUES ($1, $2, $3, $4, $5)`,
			userID2, "User Two", "test.user@example.com", "hash2", "mahasiswa")
		if err == nil {
			t.Fatal("expected error inserting duplicate case-insensitive email, got nil")
		}
	})

	t.Run("constraint: lecturer and student must be different", func(t *testing.T) {
		sameUserID := "33333333-3333-3333-3333-333333333333"
		sessionID := "44444444-4444-4444-4444-444444444444"

		_, err := db.ExecContext(ctx, `INSERT INTO users (id, full_name, email, password_hash, role) VALUES ($1, $2, $3, $4, $5)`,
			sameUserID, "User Same", "same.user@example.com", "hash", "dosen")
		if err != nil {
			t.Fatalf("failed to insert user: %v", err)
		}

		now := time.Now()
		_, err = db.ExecContext(ctx, `INSERT INTO guidance_sessions (id, lecturer_id, student_id, topic, scheduled_start_at, scheduled_end_at, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			sessionID, sameUserID, sameUserID, "Topic", now, now.Add(1*time.Hour), sameUserID)
		if err == nil {
			t.Fatal("expected error when lecturer_id == student_id, got nil")
		}
	})

	t.Run("constraint: scheduled_end_at must be after scheduled_start_at", func(t *testing.T) {
		lecturerID := "55555555-5555-5555-5555-555555555555"
		studentID := "66666666-6666-6666-6666-666666666666"
		sessionID := "77777777-7777-7777-7777-777777777777"

		_, _ = db.ExecContext(ctx, `INSERT INTO users (id, full_name, email, password_hash, role) VALUES ($1, 'Lecturer', 'lecturer@example.com', 'hash', 'dosen')`, lecturerID)
		_, _ = db.ExecContext(ctx, `INSERT INTO users (id, full_name, email, password_hash, role) VALUES ($1, 'Student', 'student@example.com', 'hash', 'mahasiswa')`, studentID)

		now := time.Now()
		// Invalid: end time before start time
		_, err = db.ExecContext(ctx, `INSERT INTO guidance_sessions (id, lecturer_id, student_id, topic, scheduled_start_at, scheduled_end_at, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			sessionID, lecturerID, studentID, "Topic", now, now.Add(-1*time.Hour), lecturerID)
		if err == nil {
			t.Fatal("expected error when scheduled_end_at <= scheduled_start_at, got nil")
		}
	})

	t.Run("constraint: negative document size is rejected", func(t *testing.T) {
		lecturerID := "55555555-5555-5555-5555-555555555555"
		studentID := "66666666-6666-6666-6666-666666666666"
		sessionID := "88888888-8888-8888-8888-888888888888"
		docID := "99999999-9999-9999-9999-999999999999"

		now := time.Now()
		_, _ = db.ExecContext(ctx, `INSERT INTO guidance_sessions (id, lecturer_id, student_id, topic, scheduled_start_at, scheduled_end_at, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			sessionID, lecturerID, studentID, "Topic Valid", now, now.Add(1*time.Hour), lecturerID)

		// Invalid: negative size_bytes
		_, err = db.ExecContext(ctx, `INSERT INTO documents (id, guidance_session_id, uploaded_by, original_filename, storage_bucket, storage_key, mime_type, size_bytes, checksum_sha256)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			docID, sessionID, studentID, "file.pdf", "bucket", "key-1", "application/pdf", -100, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
		if err == nil {
			t.Fatal("expected error inserting negative size_bytes, got nil")
		}
	})

	t.Run("constraint: orphan foreign key is rejected", func(t *testing.T) {
		orphanUserID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
		_, err := db.ExecContext(ctx, `INSERT INTO refresh_sessions (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`,
			"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", orphanUserID, "hash", time.Now().Add(1*time.Hour))
		if err == nil {
			t.Fatal("expected foreign key constraint violation inserting orphan user_id, got nil")
		}
	})
}
