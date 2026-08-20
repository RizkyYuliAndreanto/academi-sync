package migrations

import "embed"

// FS embeds all SQL migration files in this package directory.
//
//go:embed *.sql
var FS embed.FS
