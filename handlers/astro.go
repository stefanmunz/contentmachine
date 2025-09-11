package handlers

import (
	"distribute/config"
	"distribute/models"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// writeToBlog writes content to a specific blog
func writeToBlog(blogConfig config.BlogConfig, blogName string, content *models.Content, dryRun bool) error {
	// Skip if blog is not configured
	if blogConfig.ContentPath == "" {
		return nil
	}
	
	// Generate filename from title
	filename := generateFilenameFromTitle(content.Metadata.Title)
	destPath := filepath.Join(blogConfig.ContentPath, filename)
	
	// Check for banner image
	sourceDir := filepath.Dir(content.OriginalPath)
	bannerSourcePath := filepath.Join(sourceDir, "banner.jpg")
	bannerFileName := ""
	
	// Check if banner exists
	if _, err := os.Stat(bannerSourcePath); err == nil {
		// Generate banner filename based on title
		bannerFileName = strings.TrimSuffix(filename, ".md") + "-banner.jpg"
	}
	
	if dryRun {
		fmt.Printf("📁 %s BLOG:\n", strings.ToUpper(blogName))
		fmt.Printf("Would create: %s\n", destPath)
		if bannerFileName != "" {
			fmt.Printf("Would copy banner: %s → %s\n", bannerSourcePath, filepath.Join(blogConfig.ContentPath, bannerFileName))
		}
		return nil
	}
	
	// Copy banner image if it exists
	if bannerFileName != "" {
		bannerDestPath := filepath.Join(blogConfig.ContentPath, bannerFileName)
		if err := copyFile(bannerSourcePath, bannerDestPath); err != nil {
			log.Printf("WARNING: Failed to copy banner image to %s: %v", blogName, err)
			bannerFileName = "" // Clear it so we don't include it in the markdown
		}
	}
	
	// Build the blog post content with only frontmatter and main content
	// Add author based on blog type
	author := ""
	if blogName == "OnTree" {
		author = "onTree Team"
	} else if blogName == "Personal" {
		author = "Stefan Munz"
	}
	blogContent := buildBlogContent(content, bannerFileName, author)
	
	// Write to destination file
	if err := os.WriteFile(destPath, []byte(blogContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s blog file: %w", blogName, err)
	}
	
	log.Printf("INFO: %s blog post created at %s", blogName, destPath)
	if bannerFileName != "" {
		log.Printf("INFO: Banner image copied to %s as %s", blogName, bannerFileName)
	}
	
	return nil
}

func HandleAstroPost(cfg *config.Config, content *models.Content, dryRun bool) error {
	// Write to personal blog
	if err := writeToBlog(cfg.PersonalBlog, "Personal", content, dryRun); err != nil {
		log.Printf("ERROR: Failed to write to personal blog: %v", err)
		// Continue with OnTree blog even if personal blog fails
	}
	
	// Write to OnTree blog if configured
	if cfg.OnTreeBlog.ContentPath != "" {
		if err := writeToBlog(cfg.OnTreeBlog, "OnTree", content, dryRun); err != nil {
			log.Printf("ERROR: Failed to write to OnTree blog: %v", err)
		}
	}
	
	if dryRun {
		fmt.Printf("\n")
	}
	
	return nil
}

// buildBlogContent creates the blog post content with frontmatter and main content only
func buildBlogContent(content *models.Content, bannerFileName string, author string) string {
	var builder strings.Builder
	
	// Add frontmatter
	builder.WriteString("---\n")
	
	// Add author if provided (for OnTree blog)
	if author != "" {
		builder.WriteString(fmt.Sprintf("author: %s\n", author))
	}
	
	// Use current time in CEST/CET timezone
	location, err := time.LoadLocation("Europe/Berlin") // CEST/CET
	if err != nil {
		// Fallback to UTC if timezone not found
		location = time.UTC
	}
	currentTime := time.Now().In(location)
	builder.WriteString(fmt.Sprintf("pubDatetime: %s\n", currentTime.Format(time.RFC3339)))
	
	builder.WriteString(fmt.Sprintf("title: %q\n", content.Metadata.Title))
	
	// Generate postSlug from title
	postSlug := generateFilenameFromTitle(content.Metadata.Title)
	postSlug = strings.TrimSuffix(postSlug, ".md")
	builder.WriteString(fmt.Sprintf("postSlug: %s\n", postSlug))
	
	// Add featured and draft fields (for OnTree compatibility)
	builder.WriteString("featured: false\n")
	builder.WriteString("draft: false\n")
	
	// Add description (first paragraph of thought piece)
	description := content.ThoughtPiece
	if idx := strings.Index(description, "\n"); idx > 0 {
		description = description[:idx]
	}
	if len(description) > 160 {
		description = description[:157] + "..."
	}
	builder.WriteString(fmt.Sprintf("description: %q\n", description))
	
	// Add tags
	if len(content.Metadata.Tags) > 0 {
		builder.WriteString("tags:\n")
		for _, tag := range content.Metadata.Tags {
			builder.WriteString(fmt.Sprintf("  - %s\n", tag))
		}
	}
	
	builder.WriteString("---\n\n")
	
	// Add banner image if it exists
	if bannerFileName != "" {
		builder.WriteString(fmt.Sprintf("![%s](./%s)\n\n", content.Metadata.Title, bannerFileName))
	}
	
	// Add the main content (thought piece only)
	builder.WriteString(content.ThoughtPiece)
	
	return builder.String()
}

// generateFilenameFromTitle creates a clean filename from the post title
func generateFilenameFromTitle(title string) string {
	// Convert to lowercase
	filename := strings.ToLower(title)
	
	// Remove special characters except spaces and alphanumeric
	re := regexp.MustCompile(`[^a-z0-9\s]+`)
	filename = re.ReplaceAllString(filename, "")
	
	// Replace multiple spaces with single space
	re = regexp.MustCompile(`\s+`)
	filename = re.ReplaceAllString(filename, " ")
	
	// Trim spaces and replace with dashes
	filename = strings.TrimSpace(filename)
	filename = strings.ReplaceAll(filename, " ", "-")
	
	// Add .md extension
	return filename + ".md"
}

// copyFile copies a file from source to destination
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}