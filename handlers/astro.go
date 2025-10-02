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

	// Generate folder name from title
	folderName := generateFilenameFromTitle(content.Metadata.Title)
	folderName = strings.TrimSuffix(folderName, ".mdx") // Remove .mdx extension for folder name

	// Extract year from publish date or use current year
	year := time.Now().Year()
	if content.Metadata.PublishDate != "" {
		// Try to parse the publish date to get the year
		if t, err := time.Parse(time.RFC3339, content.Metadata.PublishDate); err == nil {
			year = t.Year()
		} else if t, err := time.Parse("2006-01-02T15:04:05-07:00", content.Metadata.PublishDate); err == nil {
			year = t.Year()
		}
	}

	// Create the full path: content/blog/2025/post-folder/index.mdx
	postDir := filepath.Join(blogConfig.ContentPath, fmt.Sprintf("%d", year), folderName)
	destPath := filepath.Join(postDir, "index.mdx")

	// Check for banner image (try both .jpg and .png)
	sourceDir := filepath.Dir(content.OriginalPath)
	bannerSourcePath := ""
	bannerFileName := ""

	// Check if banner.jpg exists
	if _, err := os.Stat(filepath.Join(sourceDir, "banner.jpg")); err == nil {
		bannerSourcePath = filepath.Join(sourceDir, "banner.jpg")
		bannerFileName = "banner.jpg"
	} else if _, err := os.Stat(filepath.Join(sourceDir, "banner.png")); err == nil {
		// Check if banner.png exists
		bannerSourcePath = filepath.Join(sourceDir, "banner.png")
		bannerFileName = "banner.png"
	}

	// Find all image files in the source directory (except banner.jpg/banner.png which is handled separately)
	inlineImages := []string{}
	files, err := os.ReadDir(sourceDir)
	if err == nil {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			name := file.Name()
			ext := strings.ToLower(filepath.Ext(name))
			// Check if it's an image file (excluding the banner)
			if (ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp") &&
				!strings.HasPrefix(name, "banner.") {
				inlineImages = append(inlineImages, name)
			}
		}
	}

	if dryRun {
		fmt.Printf("📁 %s BLOG:\n", strings.ToUpper(blogName))
		fmt.Printf("Would create directory: %s\n", postDir)
		fmt.Printf("Would create: %s\n", destPath)
		if bannerFileName != "" {
			fmt.Printf("Would copy banner: %s → %s\n", bannerSourcePath, filepath.Join(postDir, bannerFileName))
		}
		// Show inline images that would be copied
		for _, img := range inlineImages {
			imgSource := filepath.Join(sourceDir, img)
			fmt.Printf("Would copy image: %s → %s\n", imgSource, filepath.Join(postDir, img))
		}
		return nil
	}

	// Create the post directory
	if err := os.MkdirAll(postDir, 0755); err != nil {
		return fmt.Errorf("failed to create post directory: %w", err)
	}

	// Copy banner image if it exists
	if bannerFileName != "" {
		bannerDestPath := filepath.Join(postDir, bannerFileName)
		if err := copyFile(bannerSourcePath, bannerDestPath); err != nil {
			log.Printf("WARNING: Failed to copy banner image to %s: %v", blogName, err)
			bannerFileName = "" // Clear it so we don't include it in the markdown
		}
	}

	// Copy inline images
	for _, img := range inlineImages {
		imgSource := filepath.Join(sourceDir, img)
		imgDest := filepath.Join(postDir, img)
		if err := copyFile(imgSource, imgDest); err != nil {
			log.Printf("WARNING: Failed to copy image %s to %s: %v", img, blogName, err)
		} else {
			log.Printf("INFO: Copied image %s to %s blog", img, blogName)
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

	// Process the thought piece to use relative image paths for the blog
	// First, remove the banner image from the thought piece to avoid duplication
	thoughtPieceWithoutBanner := content.ThoughtPiece
	// Remove the first image (banner) from the thought piece
	bannerImagePattern := regexp.MustCompile(`(?m)^!\[.*?\]\(.*?banner\.png\)\s*\n*`)
	thoughtPieceWithoutBanner = bannerImagePattern.ReplaceAllString(thoughtPieceWithoutBanner, "")

	processedThoughtPiece := processImagePathsForBlog(thoughtPieceWithoutBanner)

	blogContent := buildBlogContentWithProcessedContent(content, processedThoughtPiece, bannerFileName, author)

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

// processImagePathsForBlog converts image paths to Astro Image components for blog posts
func processImagePathsForBlog(content string) string {
	// Track unique images for imports
	imageImports := make(map[string]string) // filename -> importName

	// Use regex to find all markdown image references
	re := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	// First pass: collect all images and generate import names
	re.ReplaceAllStringFunc(content, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		imagePath := submatches[2]
		filename := filepath.Base(imagePath)

		// Check if it's an image file based on extension
		ext := strings.ToLower(filepath.Ext(filename))
		imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg"}
		isImage := false
		for _, validExt := range imageExts {
			if ext == validExt {
				isImage = true
				break
			}
		}

		if isImage && imageImports[filename] == "" {
			// Generate import name from filename (e.g., "image.png" -> "imageImg")
			baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
			// Clean the base name to be a valid JS identifier
			cleanName := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(baseName, "")
			importName := cleanName + "Img"
			imageImports[filename] = importName
		}

		return match
	})

	// Build imports section
	var importsBuilder strings.Builder
	if len(imageImports) > 0 {
		// Add Image import if not already present
		if !strings.Contains(content, "import { Image } from \"astro:assets\"") {
			importsBuilder.WriteString("import { Image } from \"astro:assets\";\n")
		}

		// Add individual image imports
		for filename, importName := range imageImports {
			importsBuilder.WriteString(fmt.Sprintf("import %s from \"./%s\";\n", importName, filename))
		}
		importsBuilder.WriteString("\n")
	}

	// Second pass: replace markdown images with Image components
	result := re.ReplaceAllStringFunc(content, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		altText := submatches[1]
		imagePath := submatches[2]
		filename := filepath.Base(imagePath)

		// Check if we have an import for this image
		if importName, exists := imageImports[filename]; exists {
			return fmt.Sprintf("<Image src={%s} alt=\"%s\" widths={[400, 800, 1200]} sizes=\"(max-width: 800px) 100vw, 800px\" />", importName, altText)
		}

		return match
	})

	// Prepend imports to the content
	if len(imageImports) > 0 {
		result = importsBuilder.String() + result
	}

	return result
}

// buildBlogContentWithProcessedContent creates the blog post content with frontmatter and processed content
func buildBlogContentWithProcessedContent(content *models.Content, processedThoughtPiece string, bannerFileName string, author string) string {
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
	postSlug = strings.TrimSuffix(postSlug, ".mdx")
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

	// Add Astro Image import for banner if it exists
	if bannerFileName != "" {
		builder.WriteString("import { Image } from \"astro:assets\";\n")
		// Create import name from banner filename (e.g., banner.png -> bannerImg)
		importName := strings.TrimSuffix(bannerFileName, filepath.Ext(bannerFileName)) + "Img"
		builder.WriteString(fmt.Sprintf("import %s from \"./%s\";\n\n", importName, bannerFileName))
		// Use widths prop for responsive images - Astro will generate multiple sizes
		builder.WriteString(fmt.Sprintf("<Image\n  src={%s}\n  alt=\"%s\"\n  widths={[400, 800, 1200]}\n  sizes=\"(max-width: 800px) 100vw, 800px\"\n/>\n\n", importName, content.Metadata.Title))
	}

	// Add the processed content with relative paths
	builder.WriteString(processedThoughtPiece)

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

	// Add .mdx extension
	return filename + ".mdx"
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
