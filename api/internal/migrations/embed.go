package migrations

import (
	"fmt"
	"io/fs"

	migrationfiles "auth/migrations"

	"ariga.io/atlas/sql/migrate"
)

// EmbeddedDir returns the Atlas migration directory embedded in the API binary.
func EmbeddedDir() (migrate.Dir, error) {
	dir := migrate.OpenMemDir("auth")
	dir.Reset()

	entries, err := fs.ReadDir(migrationfiles.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := migrationfiles.FS.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read embedded migration %q: %w", entry.Name(), err)
		}
		if err := dir.WriteFile(entry.Name(), data); err != nil {
			return nil, fmt.Errorf("copy embedded migration %q: %w", entry.Name(), err)
		}
	}
	return dir, nil
}
