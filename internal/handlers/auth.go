package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"net/url"


	"github.com/eduardobachmann/back-portfolio/internal/config"
	"github.com/eduardobachmann/back-portfolio/internal/models"
	"github.com/eduardobachmann/back-portfolio/internal/services"
)

type AuthHandler struct {
	cfg     *config.Config
	authSvc *services.AuthService
}

func NewAuthHandler(cfg *config.Config, authSvc *services.AuthService) *AuthHandler {
	return &AuthHandler{
		cfg:     cfg,
		authSvc: authSvc,
	}
}

// HandleLinkedInLogin redirects to LinkedIn Auth page
func (h *AuthHandler) HandleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	scope := "openid profile email"
	
	clientID := h.cfg.LinkedInClientID
	if clientID == "" {
		log.Println("WARNING: LinkedIn Client ID is empty in configuration")
	}

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", clientID)
	params.Add("redirect_uri", h.cfg.LinkedInRedirectURI)
	params.Add("state", "random_state_string_for_security")
	params.Add("scope", scope)

	authURL := "https://www.linkedin.com/oauth/v2/authorization?" + params.Encode()
	
	log.Printf("Redirecting to LinkedIn Auth: %s\n", authURL)
	
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}


// HandleLinkedInCallback handles the redirect from LinkedIn
func (h *AuthHandler) HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found in request", http.StatusBadRequest)
		return
	}

	// Exchange Code for Access Token
	token, err := h.exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Token exchange failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Get User URN
	user, err := h.getLinkedInUser(token.AccessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}

	// Save to DB
	if err := h.authSvc.SaveLinkedInToken(r.Context(), token, user.Sub); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save token: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully linked LinkedIn account: %s", user.Name)
}

func (h *AuthHandler) exchangeCodeForToken(code string) (*models.TokenResponse, error) {
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("client_id", h.cfg.LinkedInClientID)
	form.Add("client_secret", h.cfg.LinkedInClientSecret)
	form.Add("redirect_uri", h.cfg.LinkedInRedirectURI)

	resp, err := http.PostForm("https://www.linkedin.com/oauth/v2/accessToken", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("linkedin error: %v", errResp)
	}

	var token models.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&token)
	return &token, err
}

func (h *AuthHandler) getLinkedInUser(token string) (*models.LinkedInUser, error) {
	req, _ := http.NewRequest("GET", "https://api.linkedin.com/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user models.LinkedInUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	return &user, err
}
