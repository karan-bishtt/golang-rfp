package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/karan-bishtt/user-service/config"
	"github.com/karan-bishtt/user-service/internal/models"
)

type AuthService struct {
	baseURL string
	client  *http.Client
}

type AuthServiceResponse struct {
	Status  int                   `json:"status"`
	Message string                `json:"message"`
	Data    []models.UserResponse `json:"data"`
}

func NewAuthService() *AuthService {
	cfg := config.Load()
	return &AuthService{
		baseURL: cfg.AuthServiceURL,
		client:  &http.Client{},
	}
}

// GetVendors fetches all vendor users from auth-service
func (as *AuthService) GetVendors() ([]models.UserResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users?role=vendor", as.baseURL)

	resp, err := as.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vendors: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status: %d", resp.StatusCode)
	}

	var response AuthServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return response.Data, nil
}

// GetUserByID fetches a specific user from auth-service
func (as *AuthService) GetUserByID(userID uint) (*models.UserResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users/%d", as.baseURL, userID)

	resp, err := as.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user not found or auth service error")
	}

	var response struct {
		Status  int                 `json:"status"`
		Message string              `json:"message"`
		Data    models.UserResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response.Data, nil
}
