package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bimbingan-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates new UUID v4 when header is missing", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RequestID())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		reqID := resp.Header().Get("X-Request-ID")
		if reqID == "" {
			t.Fatal("expected X-Request-ID header to be populated")
		}
		if len(reqID) < 20 {
			t.Fatalf("expected valid UUID v4 string, got: %s", reqID)
		}
	})

	t.Run("preserves valid incoming X-Request-ID header", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RequestID())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		customID := "req-12345-abcde"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", customID)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		reqID := resp.Header().Get("X-Request-ID")
		if reqID != customID {
			t.Fatalf("expected X-Request-ID to be %s, got: %s", customID, reqID)
		}
	})

	t.Run("accepts X-Request-ID with exactly 64 characters", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RequestID())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		exact64ID := strings.Repeat("a", 64)
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", exact64ID)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		reqID := resp.Header().Get("X-Request-ID")
		if reqID != exact64ID {
			t.Fatalf("expected 64-character X-Request-ID to be kept, got: %s", reqID)
		}
	})

	t.Run("replaces X-Request-ID longer than 64 characters", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RequestID())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		tooLongID := strings.Repeat("a", 65)
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", tooLongID)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		reqID := resp.Header().Get("X-Request-ID")
		if reqID == tooLongID {
			t.Fatalf("expected X-Request-ID longer than 64 chars to be replaced")
		}
	})

	t.Run("replaces X-Request-ID containing invalid chars (spaces, newlines, unicode, html)", func(t *testing.T) {
		invalidIDs := []string{
			"invalid id with spaces",
			"invalid\nid\nwith\nnewlines",
			"invalid-unicode-🚀-id",
			"<script>alert(1)</script>",
			"id;drop table users;",
		}

		for _, invalidID := range invalidIDs {
			router := gin.New()
			router.Use(middleware.RequestID())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Request-ID", invalidID)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			reqID := resp.Header().Get("X-Request-ID")
			if reqID == invalidID {
				t.Fatalf("expected invalid X-Request-ID %q to be replaced, but it was kept", invalidID)
			}
		}
	})
}
