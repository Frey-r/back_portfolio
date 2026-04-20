package services

import (
	"context"
	"database/sql"
	"time"


	"github.com/eduardobachmann/back-portfolio/internal/database"
	"github.com/eduardobachmann/back-portfolio/internal/models"
)

type AuthService struct {
	db *database.DB
}

func NewAuthService(db *database.DB) *AuthService {
	return &AuthService{db: db}
}

// SaveLinkedInToken persists the LinkedIn token and URN in the database
func (s *AuthService) SaveLinkedInToken(ctx context.Context, token *models.TokenResponse, urn string) error {
	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	
	query := `
		INSERT INTO auth_tokens (provider, access_token, refresh_token, user_urn, expires_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(provider) DO UPDATE SET
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			user_urn = excluded.user_urn,
			expires_at = excluded.expires_at,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.ExecContext(ctx, query, "linkedin", token.AccessToken, token.RefreshToken, urn, expiresAt)
	return err
}

// GetLinkedInToken retrieves the token from the database
func (s *AuthService) GetLinkedInToken(ctx context.Context) (*models.AuthToken, error) {
	query := `SELECT access_token, refresh_token, user_urn, expires_at FROM auth_tokens WHERE provider = 'linkedin'`
	row := s.db.QueryRowContext(ctx, query)
	
	var t models.AuthToken
	var expiresAt string
	err := row.Scan(&t.AccessToken, &t.RefreshToken, &t.UserURN, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil // No token found
	}
	if err != nil {
		return nil, err
	}

	// Try to parse the datetime (SQLite stores as string)
	t.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)
	t.Provider = "linkedin"
	
	return &t, nil
}
