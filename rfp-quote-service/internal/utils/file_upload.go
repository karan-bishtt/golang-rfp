package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadFile handles file uploads
func UploadFile(file multipart.File, header *multipart.FileHeader, uploadDir string) (string, error) {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Validate file type (optional - add your allowed types)
	allowedTypes := []string{".pdf", ".doc", ".docx", ".txt", ".jpg", ".jpeg", ".png"}
	ext := strings.ToLower(filepath.Ext(header.Filename))

	isAllowed := false
	for _, allowedType := range allowedTypes {
		if ext == allowedType {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("file type %s is not allowed", ext)
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, header.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// Create the file
	dst, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	// Copy the uploaded file to destination
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return filename, nil
}
