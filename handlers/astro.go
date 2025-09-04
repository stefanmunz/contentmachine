package handlers

import (
	"distribute/config"
	"distribute/models"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func HandleAstroPost(cfg *config.Config, content *models.Content, dryRun bool) error {
	// Get the filename from the original path
	filename := filepath.Base(content.OriginalPath)
	destPath := filepath.Join(cfg.AstroContentPath, filename)
	
	if dryRun {
		fmt.Printf("📁 ASTRO BLOG:\n")
		fmt.Printf("Would copy: %s → %s\n\n", content.OriginalPath, destPath)
		return nil
	}
	
	// Read source file
	sourceFile, err := os.Open(content.OriginalPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()
	
	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()
	
	// Copy file content
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	
	log.Printf("INFO: Blog post copied to %s", destPath)
	return nil
}