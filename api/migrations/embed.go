package migrations

import "embed"

// FS contains the Atlas migration directory used at API startup.
//
//go:embed *.sql atlas.sum
var FS embed.FS
