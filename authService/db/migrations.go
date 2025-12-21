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
			CREATE TABLE IF NOT EXISTS users (
				id UUID PRIMARY KEY DEFAULT uuidv4(),
				username VARCHAR(50) UNIQUE NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS auth (
				user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
				login VARCHAR(255) UNIQUE NOT NULL,
				password VARCHAR(255) NOT NULL,
				jwt TEXT
			);
			CREATE TABLE IF NOT EXISTS groups (
				id UUID PRIMARY KEY DEFAULT uuidv4(),
				name VARCHAR(50) NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS user_group (
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
				PRIMARY KEY (user_id, group_id)
			);
			CREATE INDEX IF NOT EXISTS idx_user_group_user_id ON user_group(user_id);
			CREATE INDEX IF NOT EXISTS idx_user_group_group_id ON user_group(group_id);
			CREATE INDEX IF NOT EXISTS idx_auth_jwt ON auth(jwt);
		`,
	},
	{
		Name: "create_open_group",
		SQL: `
			INSERT INTO groups (id, name, created_at) 
        VALUES (
            '00000000-0000-0000-0000-000000000000',
            'Все пользователи',
            CURRENT_TIMESTAMP
        )
        ON CONFLICT (id) DO NOTHING;
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
