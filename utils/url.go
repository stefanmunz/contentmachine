package utils

import (
	"path/filepath"
	"strings"
)

// BuildBlogURL constructs the full blog post URL from the base URL and file path
func BuildBlogURL(baseURL, filePath string) string {
	// Get the filename without extension
	filename := filepath.Base(filePath)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	
	// Ensure base URL doesn't end with a slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	// Build the full URL
	return baseURL + "/blog/" + filename + "/"
}