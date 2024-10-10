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

	query = `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`
	_, err = conn.Exec(ctx, query)
	if err != nil {
		return err
	}

	query = `
		CREATE TABLE IF NOT EXISTS transactions (
    	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		user_id VARCHAR(255) NOT NULL,
		amount FLOAT NOT NULL,
		currency VARCHAR(255) NOT NULL,
		done BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		processed_at TIMESTAMP,
		processing_time INTERVAL
		); 
	`
	_, err = conn.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
