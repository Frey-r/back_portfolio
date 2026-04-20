package services

import (
	"context"
	"fmt"
	"time"

	"github.com/eduardobachmann/back-portfolio/internal/database"
	"github.com/eduardobachmann/back-portfolio/internal/models"
)

type ContactService struct {
	db *database.DB
}

func NewContactService(db *database.DB) *ContactService {
	return &ContactService{db: db}
}

// Create stores a new contact submission in the database
func (s *ContactService) Create(ctx context.Context, req *models.ContactRequest, ip, userAgent string) (int64, error) {
	query := `
		INSERT INTO contacts (name, email, message, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := s.db.ExecContext(ctx, query, req.Name, req.Email, req.Message, ip, userAgent)
	if err != nil {
		return 0, fmt.Errorf("failed to insert contact: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert id: %w", err)
	}

	return id, nil
}

// List retrieved contact submissions (for optional admin panel later)
func (s *ContactService) List(ctx context.Context, limit, offset int) ([]*models.ContactRecord, error) {
	query := `
		SELECT id, name, email, message, status, created_at, updated_at
		FROM contacts
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*models.ContactRecord
	for rows.Next() {
		var c models.ContactRecord
		var createdAt, updatedAt string
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Message, &c.Status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		
		contacts = append(contacts, &c)
	}

	return contacts, nil
}
