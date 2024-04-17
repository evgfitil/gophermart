package database

import "context"

func (db *DBStorage) CreateUser(ctx context.Context, username string, passwordHash string) error {
	_, err := db.conn.ExecContext(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2)`, username, passwordHash)
	return err
}

func (db *DBStorage) IsUserUnique(ctx context.Context, username string) (bool, error) {
	var userExists bool
	row := db.conn.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username)
	if err := row.Scan(&userExists); err != nil {
		return false, err
	}
	return !userExists, nil
}
