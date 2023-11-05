package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func runMigrations() {
	log.Println("Running migrations (if applicable)...")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", getEnv("DB_USER"), getEnv("DB_PASSWORD"), getEnv("DB_NAME"), getEnv("DB_HOST"), getEnv("DB_PORT"))

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Panic("error opening database for migration: ", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Panic("error creating database driver for migration: ", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx",
		driver,
	)
	if err != nil {
		log.Panic("error creating new migrate instance: ", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to run.")
		} else {
			log.Panic("error running migration: ", err)
		}
	}
	log.Println("Completed migrations run.")
}
