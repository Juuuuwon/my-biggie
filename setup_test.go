package main

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

// SetupTestDatabase automatically creates testing schemas and/or tables
// for external relational databases. It supports "mysql", "postgres", and "redshift".
// For MySQL: Creates a table "biggie_test_table" in the current database.
// For PostgreSQL: Creates a schema "biggie_test_schema" and a table "biggie_test_table" within it.
// For Redshift: Creates a table "biggie_test_table" in the default schema.
func SetupTestDatabase(dbType string, db *sql.DB) error {
	switch dbType {
	case "mysql":
		query := `
			CREATE TABLE IF NOT EXISTS biggie_test_table (
				id INT AUTO_INCREMENT PRIMARY KEY,
				value VARCHAR(255) NOT NULL
			);
		`
		if _, err := db.Exec(query); err != nil {
			log("failed to create test table for MySQL", zap.Error(err))
			return err
		}
		log("MySQL test table created or already exists")
		return nil

	case "postgres":
		// Create schema if it does not exist.
		if _, err := db.Exec(`CREATE SCHEMA IF NOT EXISTS biggie_test_schema;`); err != nil {
			log("failed to create test schema for PostgreSQL", zap.Error(err))
			return err
		}
		query := `
			CREATE TABLE IF NOT EXISTS biggie_test_schema.biggie_test_table (
				id SERIAL PRIMARY KEY,
				value TEXT NOT NULL
			);
		`
		if _, err := db.Exec(query); err != nil {
			log("failed to create test table for PostgreSQL", zap.Error(err))
			return err
		}
		log("PostgreSQL test schema and table created or already exists")
		return nil

	case "redshift":
		// Redshift uses similar syntax to PostgreSQL; here we create a table in the default schema.
		query := `
			CREATE TABLE IF NOT EXISTS biggie_test_table (
				id INT IDENTITY(1,1) PRIMARY KEY,
				value VARCHAR(255) NOT NULL
			);
		`
		if _, err := db.Exec(query); err != nil {
			log("failed to create test table for Redshift", zap.Error(err))
			return err
		}
		log("Redshift test table created or already exists")
		return nil

	default:
		return fmt.Errorf("unsupported dbType: %s", dbType)
	}
}
