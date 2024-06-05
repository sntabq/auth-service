package storage

import (
	"auth-service/internal/data/models"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func (s *AuthStorage) SaveToken(ctx context.Context, tokenPlainText string, userId int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"
	tokenHash := sha256.Sum256([]byte(tokenPlainText))
	fail := func(e error) error {
		return fmt.Errorf("%s: %v", op, e)
	}
	stmt, err := s.db.Prepare(`
								INSERT INTO auth.tokens(hash, user_id, expiry) 
								VALUES ($1, $2, $3)
								`)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, fmt.Errorf("%s: %w", op, ErrTokenNotSaved)
		default:
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fail(err)
	}
	defer tx.Rollback()

	row := stmt.QueryRowContext(ctx, tokenHash[:], userId, time.Now().Add(time.Hour))

	if row.Err() != nil {
		return false, fail(err)
	}
	return true, nil
}

func (s *AuthStorage) IsAuthenticated(ctx context.Context, tokenPlainText string) (bool, error) {
	const op = "storage.sqlite.IsAdmin"
	tokenHash := sha256.Sum256([]byte(tokenPlainText))
	stmt, err := s.db.Prepare(`
								SELECT auth.users.id,
								       auth.users.username,
								       auth.users.email,
								       auth.users.password_hash,
								       auth.users.user_role,
								       auth.users.activated
								           FROM auth.users
								    INNER JOIN auth.tokens t ON
								        auth.users.id = t.user_id
								         WHERE t.hash = $1
								           AND t.expiry > $2
								`)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		default:
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}

	row := stmt.QueryRowContext(ctx, tokenHash[:], time.Now())

	if errors.Is(row.Err(), sql.ErrNoRows) {
		return false, nil
	}

	var user models.User
	err = row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.Hash,
		&user.Role,
		&user.Activated,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (s *AuthStorage) CheckTokens() {
	for {
		queryToGetExpiredUserIds := `
			SELECT u.user_id 
			FROM auth.tokens u 
			WHERE expiry < now()`

		rows, err := s.db.Query(queryToGetExpiredUserIds)
		if err != nil {
			log.Printf("failed to get expired users")
			return
		}

		var userInfos []*models.User
		for rows.Next() {
			var userInfo models.User
			err := rows.Scan(&userInfo.ID)
			if err != nil {
				log.Printf("failed to scan expired user")
				return
			}
			userInfos = append(userInfos, &userInfo)
		}

		queryToDeleteExpiredTokens := `
				DELETE FROM auth.tokens
				WHERE user_id = $1`

		for _, ui := range userInfos {
			_, err := s.db.Exec(queryToDeleteExpiredTokens, ui.ID)
			if err != nil {
				log.Printf("failed to delete expired user with id = %d", ui.ID)
				return
			}
		}

		time.Sleep(time.Minute * 20)
	}
}
