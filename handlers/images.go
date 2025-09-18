package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ImageUploader handles copying images to the blog repository
type ImageUploader struct {
	BlogRepoPath string
	BaseURL      string
}

// NewImageUploader creates a new image uploader instance
func NewImageUploader(blogRepoPath, baseURL string) *ImageUploader {
	return &ImageUploader{
		BlogRepoPath: blogRepoPath,
		BaseURL:      baseURL,
	}
}

// UploadImage copies an image from the content directory to the blog repository
// Returns the public URL of the uploaded image
func (u *ImageUploader) UploadImage(sourcePath string, issueNumber string) (string, error) {
	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", fmt.Errorf("source image not found: %s", sourcePath)
	}

	// Create the newsletter subdirectory in the blog's public/images
	// First ensure public and images directories exist
	publicDir := filepath.Join(u.BlogRepoPath, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create public directory: %w", err)
	}

	imagesDir := filepath.Join(publicDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create images directory: %w", err)
	}

	newsletterDir := filepath.Join(imagesDir, "newsletter", issueNumber)
	if err := os.MkdirAll(newsletterDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create newsletter directory: %w", err)
	}

	// Get the filename from source path
	filename := filepath.Base(sourcePath)

	// Destination path in the blog repository
	destPath := filepath.Join(newsletterDir, filename)

	// Copy the file
	if err := copyImageFile(sourcePath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy image: %w", err)
	}

	// Return the public URL
	publicURL := fmt.Sprintf("%s/images/newsletter/%s/%s", u.BaseURL, issueNumber, filename)
	log.Printf("INFO: Image uploaded: %s -> %s", sourcePath, publicURL)

	return publicURL, nil
}

// ProcessContentImages finds and uploads all local images in the content directory
// Returns a map of local paths to public URLs
func (u *ImageUploader) ProcessContentImages(contentDir string, issueNumber string) (map[string]string, error) {
	imageMap := make(map[string]string)

	// Common image extensions
	extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}

	// Walk through the content directory
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's an image file
		ext := strings.ToLower(filepath.Ext(path))
		isImage := false
		for _, validExt := range extensions {
			if ext == validExt {
				isImage = true
				break
			}
		}

		if !isImage {
			return nil
		}

		// Upload the image
		publicURL, err := u.UploadImage(path, issueNumber)
		if err != nil {
			log.Printf("WARNING: Failed to upload image %s: %v", path, err)
			return nil // Continue with other images
		}

		// Store the mapping (relative path from content dir)
		relPath, _ := filepath.Rel(contentDir, path)
		imageMap[relPath] = publicURL

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to process images: %w", err)
	}

	return imageMap, nil
}

// CopyNewsletterAssets copies the avatar and banner images to the issue folder
// Returns the public URLs for both images
func (u *ImageUploader) CopyNewsletterAssets(contentMachineRoot string, issueNumber string) (avatarURL string, bannerURL string, err error) {
	// Path to the content/images directory
	imagesDir := filepath.Join(contentMachineRoot, "content", "images")

	// Copy avatar.jpg
	avatarSource := filepath.Join(imagesDir, "avatar.jpg")
	if _, err := os.Stat(avatarSource); err == nil {
		avatarURL, err = u.UploadImage(avatarSource, issueNumber)
		if err != nil {
			log.Printf("WARNING: Failed to copy avatar.jpg: %v", err)
		}
	}

	// Copy newsletter_banner.png
	bannerSource := filepath.Join(imagesDir, "newsletter_banner.png")
	if _, err := os.Stat(bannerSource); err == nil {
		bannerURL, err = u.UploadImage(bannerSource, issueNumber)
		if err != nil {
			log.Printf("WARNING: Failed to copy newsletter_banner.png: %v", err)
		}
	}

	return avatarURL, bannerURL, nil
}

// copyImageFile copies a file from src to dst
func copyImageFile(src, dst string) error {
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
	if err != nil {
		return err
	}

	// Sync to ensure the write is complete
	return destFile.Sync()
}

// DownloadImage downloads an image from a URL and saves it to the blog repository
// Returns the public URL of the saved image
func (u *ImageUploader) DownloadImage(imageURL string, issueNumber string, filename string) (string, error) {
	// Create the newsletter subdirectory in the blog's public/images
	// First ensure public and images directories exist
	publicDir := filepath.Join(u.BlogRepoPath, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create public directory: %w", err)
	}

	imagesDir := filepath.Join(publicDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create images directory: %w", err)
	}

	newsletterDir := filepath.Join(imagesDir, "newsletter", issueNumber)
	if err := os.MkdirAll(newsletterDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create newsletter directory: %w", err)
	}

	// Download the image
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	// Determine file extension if not provided
	if !strings.Contains(filename, ".") {
		contentType := resp.Header.Get("Content-Type")
		switch contentType {
		case "image/jpeg", "image/jpg":
			filename += ".jpg"
		case "image/png":
			filename += ".png"
		case "image/gif":
			filename += ".gif"
		case "image/webp":
			filename += ".webp"
		default:
			// Default to jpg if we can't determine
			filename += ".jpg"
		}
	}

	// Destination path in the blog repository
	destPath := filepath.Join(newsletterDir, filename)

	// Create the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the image data
	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	// Return the public URL
	publicURL := fmt.Sprintf("%s/images/newsletter/%s/%s", u.BaseURL, issueNumber, filename)
	log.Printf("INFO: Image downloaded: %s -> %s", imageURL, publicURL)

	return publicURL, nil
}

// ReplaceImageURLs replaces local image references with public URLs in HTML content
func ReplaceImageURLs(htmlContent string, imageMap map[string]string) string {
	result := htmlContent

	// Replace each local image path with its public URL
	for localPath, publicURL := range imageMap {
		// Handle different possible references to the image
		// e.g., "./image.jpg", "image.jpg", "../image.jpg"
		filename := filepath.Base(localPath)

		// Replace various forms of the local path
		result = strings.ReplaceAll(result, fmt.Sprintf(`src="%s"`, localPath), fmt.Sprintf(`src="%s"`, publicURL))
		result = strings.ReplaceAll(result, fmt.Sprintf(`src="./%s"`, localPath), fmt.Sprintf(`src="%s"`, publicURL))
		result = strings.ReplaceAll(result, fmt.Sprintf(`src="../%s"`, localPath), fmt.Sprintf(`src="%s"`, publicURL))
		result = strings.ReplaceAll(result, fmt.Sprintf(`src="%s"`, filename), fmt.Sprintf(`src="%s"`, publicURL))

		// Also handle markdown image syntax that might be in the content
		result = strings.ReplaceAll(result, fmt.Sprintf(`](%s)`, localPath), fmt.Sprintf(`](%s)`, publicURL))
		result = strings.ReplaceAll(result, fmt.Sprintf(`](./%s)`, localPath), fmt.Sprintf(`](%s)`, publicURL))
		result = strings.ReplaceAll(result, fmt.Sprintf(`](../%s)`, localPath), fmt.Sprintf(`](%s)`, publicURL))
		result = strings.ReplaceAll(result, fmt.Sprintf(`](%s)`, filename), fmt.Sprintf(`](%s)`, publicURL))
	}

	return result
}
