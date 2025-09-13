package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/karan-bishtt/rfp-quote-service/config"
)

type AuthService struct {
	baseURL string
	client  *http.Client
}

type AuthResponse struct {
	Status  int         `json: "status"`
	Message string      `json: "message"`
	Data    interface{} `json:"data"`
}

type VendorData struct {
	ID            uint   `json:"id"`
	Email         string `json:"email"`
	VendorDetails struct {
		CategoryID *uint `json:"category_id"`
	} `json: "vendor_details"`
}

func NewAuthService() *AuthService {
	cfg := config.Load()

	return &AuthService{
		baseURL: cfg.AuthServiceURL,
		client:  &http.Client{},
	}
}

// GetVendorEmailsByIDs fetches vendor emails by their IDs
func (as *AuthService) GetVendorEmailsByIDs(vendorIDs []uint) []string {
	var emails []string
	log.Println("started fetching vendor ids")
	for _, id := range vendorIDs {
		log.Println("vendor id", id)
		url := fmt.Sprintf("%s/api/v1/auth/users/%d", as.baseURL, id)
		resp, err := as.client.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var response struct {
				Status int `json:"status"`
				Data   struct {
					Email string `json:"email"`
					Role  string `json:"role"`
				} `json:"data"`
			}

			if json.NewDecoder(resp.Body).Decode(&response) == nil && response.Data.Role == "vendor" {
				emails = append(emails, response.Data.Email)
			}
		}
	}

	return emails
}

// GetVendorEmailsByCategory fetches all vendor emails in a specific category
func (as *AuthService) GetVendorEmailsByCategory(categoryID uint) []string {
	url := fmt.Sprintf("%s/api/v1/vendors?category_id=%d", as.baseURL, categoryID)

	resp, err := as.client.Get(url)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}
	}

	var response struct {
		Status int          `json:"status"`
		Data   []VendorData `json:"data"`
	}

	if json.NewDecoder(resp.Body).Decode(&response) != nil {
		return []string{}
	}

	var emails []string
	for _, vendor := range response.Data {
		if vendor.VendorDetails.CategoryID != nil && *vendor.VendorDetails.CategoryID == categoryID {
			emails = append(emails, vendor.Email)
		}
	}

	return emails
}
