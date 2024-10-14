package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateTables(ctx context.Context, conn pgx.Conn) error {
	var (
		query string
		err   error
	)

	// Создание расширения для генерации UUID, если еще не установлено
	query = `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`
	_, err = conn.Exec(ctx, query)
	if err != nil {
		return err
	}

	// Создание таблицы для хранения песен
	query = `
		CREATE TABLE IF NOT EXISTS songs (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			group_name VARCHAR(255) NOT NULL,
			text TEXT NOT NULL,
			link TEXT,
			release_date TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = conn.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
