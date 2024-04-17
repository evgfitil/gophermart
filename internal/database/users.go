package database

import (
	"context"
	"github.com/evgfitil/gophermart.git/internal/models"
)

func (db *DBStorage) CreateUser(ctx context.Context, username string, passwordHash string) error {
	_, err := db.conn.ExecContext(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2)`, username, passwordHash)
	return err
}

func (db *DBStorage) GetUserByUsername(ctx context.Context, username string) (string, error) {
	var storedUser models.User

	row := db.conn.QueryRowContext(ctx, "SELECT username, password_hash FROM users WHERE username = $1", username)
	if err := row.Scan(&storedUser.Username, &storedUser.Password); err != nil {
		return "", err
	}
	return storedUser.Password, nil
}

func (db *DBStorage) IsUserUnique(ctx context.Context, username string) (bool, error) {
	var userExists bool
	row := db.conn.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username)
	if err := row.Scan(&userExists); err != nil {
		return false, err
	}
	return !userExists, nil
}
