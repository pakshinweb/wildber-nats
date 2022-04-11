package migration

import (
	"database/sql"
	"embed"
	"log"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations
var migrationsFolder embed.FS

// MigrateUp start the migrations
func MigrateUp(dbDriver, dbString string) error {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationsFolder,
		Root:       "migrations",
	}

	db, err := sql.Open(dbDriver, dbString)
	if err != nil {
		return err
	}
	defer db.Close()

	n, err := migrate.Exec(db, dbDriver, migrations, migrate.Up)
	if err != nil {
		return err
	}

	log.Printf("Applied %d migrations!\n", n)
	return nil
}

// MigrateDown revert transactions
func MigrateDown(dbDriver, dbString string) error {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationsFolder,
		Root:       "migrations",
	}

	db, err := sql.Open(dbDriver, dbString)
	if err != nil {
		return err
	}

	n, err := migrate.Exec(db, dbDriver, migrations, migrate.Down)
	if err != nil {
		return err
	}

	log.Printf("Down %d migrations!\n", n)
	return nil
}
