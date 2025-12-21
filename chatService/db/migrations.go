package db

import (
	"context"
	"fmt"
	"log"
)

type Migration struct {
	Name string
	SQL  string
}

var migrations = []Migration{
	{
		Name: "init",
		SQL: `
			CREATE TABLE IF NOT EXISTS chats (
				id UUID PRIMARY KEY DEFAULT uuidv4(),
				name VARCHAR(255) NOT NULL,
				pic TEXT,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS user_chat (
				user_id UUID NOT NULL,
				chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
				PRIMARY KEY (user_id, chat_id)
			);
			CREATE TABLE IF NOT EXISTS messages (
				id SERIAL PRIMARY KEY,
				chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
				user_id UUID NOT NULL,
				text TEXT NOT NULL,
				content TEXT,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at DESC);
			CREATE INDEX IF NOT EXISTS idx_user_chat_user_id ON user_chat(user_id);
			CREATE INDEX IF NOT EXISTS idx_user_chat_chat ON user_chat(chat_id);
		`,
	},
}

func (db *Database) RunMigrations(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	for _, migration := range migrations {
		var exists bool
		err := db.Pool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM migrations WHERE name = $1)",
			migration.Name,
		).Scan(&exists)

		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", migration.Name, err)
		}

		if exists {
			log.Printf("Migration %s already applied, skipping", migration.Name)
			continue
		}

		_, err = db.Pool.Exec(ctx, migration.SQL)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
		}

		_, err = db.Pool.Exec(ctx,
			"INSERT INTO migrations (name) VALUES ($1)",
			migration.Name,
		)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Name, err)
		}

		log.Printf("Applied migration: %s", migration.Name)
	}

	log.Println("All migrations applied successfully")
	return nil
}
