package models

import "time"

// JSONResponse provides a uniform structure for all API responses
type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ContactRequest represents the incoming payload from the frontend contact form.
type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// ContactRecord represents a stored contact submission in the database.
type ContactRecord struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	IPAddress string    `json:"-"` // Not exposed in generic JSON by default
	UserAgent string    `json:"-"`
	Status    string    `json:"status"` // 'new', 'read', 'replied', etc.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LinkedInPost encapsulates the details of a LinkedIn post retrieved for the portfolio.
type LinkedInPost struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	URL      string `json:"url"`
	Likes    int    `json:"likes"`
	Comments int    `json:"comments"`
	Date     string `json:"date"`
}

// TokenResponse represents the LinkedIn OAuth token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// LinkedInUser represents the simplified LinkedIn user info from /v2/userinfo
type LinkedInUser struct {
	Sub string `json:"sub"` // This is the user URN
	Name string `json:"name"`
}

// AuthToken persisted in DB
type AuthToken struct {
	Provider     string
	AccessToken  string
	RefreshToken string
	UserURN      string
	ExpiresAt    time.Time
}

// StatusInfo models the current availability status of the professional.

type StatusInfo struct {
	Available   bool   `json:"available"`
	Message     string `json:"message"`
	LastUpdated string `json:"last_updated"`
}

