package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/urdogan0000/social/internal/db"
	"github.com/urdogan0000/social/internal/env"
	"github.com/urdogan0000/social/internal/logger"
	"github.com/urdogan0000/social/posts"
	"github.com/urdogan0000/social/users"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	logger.Init("info")
	dbAddr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable")

	logger.Logger().Info().Msg("Connecting to database...")

	gormDB, err := db.NewGORM(dbAddr)
	if err != nil {
		logger.Logger().Fatal().Err(err).Msg("Failed to connect to database")
		os.Exit(1)
	}
	logger.Logger().Info().Msg("Database connected successfully")

	logger.Logger().Info().Msg("Running database migrations...")
	if err := runMigrations(gormDB); err != nil {
		logger.Logger().Fatal().Err(err).Msg("Failed to run migrations")
		os.Exit(1)
	}

	logger.Logger().Info().Msg("Migrations completed successfully")
}

func runMigrations(db *gorm.DB) error {
	logger.Logger().Info().Msg("Migrating tables: users, posts")

	if err := db.AutoMigrate(
		&users.Model{},
		&posts.Model{},
	); err != nil {
		return err
	}

	logger.Logger().Info().Msg("All migrations applied successfully")
	return nil
}
