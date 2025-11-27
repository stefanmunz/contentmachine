package handlers

import (
	"bytes"
	"distribute/config"
	"distribute/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vanng822/go-premailer/premailer"
	"github.com/yuin/goldmark"
)

// Kit v4 API request structures
type KitBroadcastRequest struct {
	EmailTemplateID int     `json:"email_template_id"`
	Subject         string  `json:"subject"`
	Content         string  `json:"content"`
	Public          bool    `json:"public"`
	SendAt          *string `json:"send_at"` // null for draft
}

type KitBroadcastResponse struct {
	Broadcast struct {
		ID int `json:"id"`
	} `json:"broadcast"`
}

type KitEmailTemplate struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Default bool   `json:"is_default"`
}

type KitEmailTemplatesResponse struct {
	EmailTemplates []KitEmailTemplate `json:"email_templates"`
}

// GetEmailTemplates fetches available email templates from Kit v4 API
func GetEmailTemplates(cfg *config.Config) ([]KitEmailTemplate, error) {
	if cfg.KitAPIKey == "" {
		return nil, fmt.Errorf("Kit API key not configured")
	}

	req, err := http.NewRequest("GET", "https://api.kit.com/v4/email_templates", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Kit-Api-Key", cfg.KitAPIKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Kit API error (status %d): %s", resp.StatusCode, string(body))
	}

	var templatesResp KitEmailTemplatesResponse
	if err := json.Unmarshal(body, &templatesResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return templatesResp.EmailTemplates, nil
}

// CreateConvertKitDraft creates a draft broadcast using Kit v4 API
func CreateConvertKitDraft(cfg *config.Config, content *models.Content, dryRun bool) error {
	// Skip if no Kit API key configured
	if cfg.KitAPIKey == "" {
		log.Println("INFO: Skipping Kit integration - KIT_API_KEY not configured")
		return nil
	}

	// Generate rich HTML with inline styles
	richHTML := formatNewsletterContentHTML(content)

	// Upload images to both blogs if configured
	richHTML = uploadAndReplaceImagesMultiBlogs(cfg, content, richHTML, dryRun)

	if dryRun {
		fmt.Printf("📧 KIT (v4 API):\n")
		fmt.Printf("Would create draft with subject: \"%s\"\n", content.Metadata.NewsletterSubject)

		// Show available templates in dry run
		templates, err := GetEmailTemplates(cfg)
		if err == nil {
			fmt.Printf("Available email templates:\n")
			for _, tmpl := range templates {
				fmt.Printf("  - ID: %d, Name: %s", tmpl.ID, tmpl.Name)
				if tmpl.Default {
					fmt.Printf(" (default)")
				}
				fmt.Printf("\n")
			}
		}

		fmt.Printf("\nHTML Preview (first 3000 chars):\n")
		preview := richHTML
		if len(preview) > 3000 {
			preview = preview[:3000] + "...[truncated]"
		}
		fmt.Printf("%s\n\n", preview)
		return nil
	}

	// First, get available templates and find the Custom HTML template
	templates, err := GetEmailTemplates(cfg)
	if err != nil {
		return fmt.Errorf("failed to fetch email templates: %w", err)
	}

	// API doesn't support "Starting point" templates, so we need to use Classic templates
	// Look for Text only template which is Classic and supports HTML content
	templateID := 0
	for _, tmpl := range templates {
		// Use Text only template - it's marked as Classic and supports HTML
		if strings.Contains(strings.ToLower(tmpl.Name), "text only") {
			templateID = tmpl.ID
			log.Printf("INFO: Using Classic template '%s' (ID: %d)", tmpl.Name, templateID)
			break
		}
	}

	// Fallback to any template if Text only not found
	if templateID == 0 && len(templates) > 0 {
		templateID = templates[0].ID
		log.Printf("INFO: Using template '%s' (ID: %d)", templates[0].Name, templateID)
	}

	if templateID == 0 {
		return fmt.Errorf("no email templates available")
	}

	// Create the request with Kit v4 structure
	reqData := KitBroadcastRequest{
		EmailTemplateID: templateID,
		Subject:         content.Metadata.NewsletterSubject,
		Content:         richHTML,
		Public:          false,
		SendAt:          nil, // nil means draft
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the request with proper v4 authentication
	req, err := http.NewRequest("POST", "https://api.kit.com/v4/broadcasts", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Kit-Api-Key", cfg.KitAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make the API call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make Kit API call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Kit API error (status %d): %s", resp.StatusCode, string(body))
	}

	var kitResp KitBroadcastResponse
	if err := json.Unmarshal(body, &kitResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	log.Printf("INFO: Kit draft created with ID: %d", kitResp.Broadcast.ID)
	return nil
}

// uploadAndReplaceImagesMultiBlogs uploads images to both blogs and replaces URLs
func uploadAndReplaceImagesMultiBlogs(cfg *config.Config, content *models.Content, htmlContent string, dryRun bool) string {
	// Upload all images (including print items) to personal blog for newsletter hosting
	if cfg.PersonalBlog.RepoPath != "" && cfg.PersonalBlog.BaseURL != "" {
		htmlContent = uploadImagesToSingleBlog(cfg.PersonalBlog, "Personal", content, htmlContent, dryRun, true)
	}

	// Upload only banner images to OnTree blog (no print items needed)
	if cfg.OnTreeBlog.RepoPath != "" && cfg.OnTreeBlog.BaseURL != "" {
		uploadImagesToSingleBlog(cfg.OnTreeBlog, "OnTree", content, htmlContent, dryRun, false)
	}

	return htmlContent
}

// uploadImagesToSingleBlog uploads images to a specific blog
func uploadImagesToSingleBlog(blogConfig config.BlogConfig, blogName string, content *models.Content, htmlContent string, dryRun bool, downloadPrintItems bool) string {
	// Extract issue number from ContentID
	issueNumber := strings.TrimPrefix(content.Metadata.ContentID, "issue")
	if issueNumber == "" {
		log.Printf("WARNING: No issue number found for %s blog, skipping image upload", blogName)
		return htmlContent
	}

	// Get the content directory path
	contentDir := filepath.Dir(content.OriginalPath)

	// Create image uploader
	uploader := NewImageUploader(blogConfig.RepoPath, blogConfig.BaseURL)

	// Only copy avatar and banner for Personal blog (stefanmunz.com)
	if blogName == "Personal" && downloadPrintItems {
		if !dryRun {
			// Get the content machine root directory (parent of content directory)
			contentMachineRoot := filepath.Dir(filepath.Dir(contentDir))
			avatarURL, bannerURL, _ := uploader.CopyNewsletterAssets(contentMachineRoot, issueNumber)

			if avatarURL != "" || bannerURL != "" {
				log.Printf("INFO: Copied newsletter assets to issue %s folder", issueNumber)
			}

			// Replace the hardcoded URLs in HTML with issue-specific ones
			if avatarURL != "" {
				htmlContent = strings.ReplaceAll(htmlContent, "https://liquid.engineer/images/avatar.jpg", avatarURL)
			}
			if bannerURL != "" {
				htmlContent = strings.ReplaceAll(htmlContent, "https://liquid.engineer/images/banner.png", bannerURL)
			}
		} else {
			// In dry run, just show what the URLs would be
			avatarURL := fmt.Sprintf("%s/images/newsletter/%s/avatar.jpg", blogConfig.BaseURL, issueNumber)
			bannerURL := fmt.Sprintf("%s/images/newsletter/%s/newsletter_banner.png", blogConfig.BaseURL, issueNumber)
			htmlContent = strings.ReplaceAll(htmlContent, "https://liquid.engineer/images/avatar.jpg", avatarURL)
			htmlContent = strings.ReplaceAll(htmlContent, "https://liquid.engineer/images/banner.png", bannerURL)
		}
	}

	if dryRun {
		fmt.Printf("Would upload images to %s blog from: %s\n", blogName, contentDir)
		fmt.Printf("To repository: %s\n", blogConfig.RepoPath)
		// Show newsletter assets copy for Personal blog
		if blogName == "Personal" && downloadPrintItems {
			fmt.Printf("Would copy newsletter assets (avatar.jpg, newsletter_banner.png) to issue %s\n", issueNumber)
		}
		// Show print images download only if enabled for this blog
		if downloadPrintItems {
			for i, item := range content.PrintItems {
				if item.ImageURL != "" && strings.HasPrefix(item.ImageURL, "http") {
					fmt.Printf("Would download print image %d to %s: %s\n", i+1, blogName, item.ImageURL)
				}
			}
		}
		return htmlContent
	}

	// Process and upload local images
	imageMap, err := uploader.ProcessContentImages(contentDir, issueNumber)
	if err != nil {
		log.Printf("WARNING: Failed to process images: %v", err)
		return htmlContent
	}

	if len(imageMap) > 0 {
		log.Printf("INFO: Uploaded %d local images to %s blog repository", len(imageMap), blogName)
		// Replace local image URLs with public URLs
		htmlContent = ReplaceImageURLs(htmlContent, imageMap)
	}

	// Download and replace print item images only if enabled
	if downloadPrintItems {
		for i, item := range content.PrintItems {
			if item.ImageURL != "" && strings.HasPrefix(item.ImageURL, "http") {
				// Generate a filename based on the item title
				filename := fmt.Sprintf("print-item-%d", i+1)
				// Sanitize title for filename
				safeTitle := strings.ToLower(item.Title)
				safeTitle = strings.ReplaceAll(safeTitle, " ", "-")
				safeTitle = strings.ReplaceAll(safeTitle, "'", "")
				safeTitle = strings.ReplaceAll(safeTitle, "\"", "")
				if safeTitle != "" {
					filename = safeTitle
				}

				// Download the image
				publicURL, err := uploader.DownloadImage(item.ImageURL, issueNumber, filename)
				if err != nil {
					log.Printf("WARNING: Failed to download print image %d to %s: %v", i+1, blogName, err)
					continue
				}

				// Replace the old URL with the new public URL in HTML content
				// Use multiple replacement patterns to handle different cases
				oldPattern := fmt.Sprintf(`src="%s"`, item.ImageURL)
				newPattern := fmt.Sprintf(`src="%s"`, publicURL)
				htmlContent = strings.ReplaceAll(htmlContent, oldPattern, newPattern)

				// Also replace in href attributes for links
				htmlContent = strings.ReplaceAll(htmlContent, item.ImageURL, publicURL)

				log.Printf("INFO: Replaced print image URL: %s -> %s", item.ImageURL, publicURL)

				// Update the content model with the new URL for consistency
				content.PrintItems[i].ImageURL = publicURL
			}
		}
	}

	return htmlContent
}

// processImagePathsForNewsletter converts image paths to absolute URLs for the newsletter
func processImagePathsForNewsletter(thoughtPiece string, contentID string) string {
	// Extract issue number from contentID (e.g., "issue50" -> "50")
	issueNumber := strings.TrimPrefix(contentID, "issue")

	// Use regex to find all markdown image references
	re := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	result := re.ReplaceAllStringFunc(thoughtPiece, func(match string) string {
		// Extract the alt text and path from the match
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		altText := submatches[1]
		imagePath := submatches[2]

		// Skip if it's already an absolute URL
		if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
			return match
		}

		// Extract just the filename from the path
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

		if !isImage {
			return match
		}

		// Build the absolute URL for www.stefanmunz.com
		absoluteURL := fmt.Sprintf("https://www.stefanmunz.com/images/newsletter/%s/%s", issueNumber, filename)

		// Return with absolute URL
		return fmt.Sprintf("![%s](%s)", altText, absoluteURL)
	})

	return result
}

func formatNewsletterContentMarkdown(content *models.Content) string {
	var builder strings.Builder

	// Add header with title and issue number
	builder.WriteString(fmt.Sprintf("# %s\n\n", content.Metadata.Title))
	builder.WriteString(fmt.Sprintf("*The Liquid Engineer – Issue No. %s*\n\n", strings.TrimPrefix(content.Metadata.ContentID, "issue")))
	builder.WriteString("---\n\n")

	// Process the thought piece to use absolute URLs for images
	processedThoughtPiece := processImagePathsForNewsletter(content.ThoughtPiece, content.Metadata.ContentID)

	// Add the processed thought piece
	builder.WriteString(processedThoughtPiece)
	builder.WriteString("\n\n")

	// Add the links section
	builder.WriteString("## What I Learned This Week\n\n")

	for _, link := range content.Links {
		// Add the description/MyTake with the link inline
		linkText := "LINK"
		if strings.Contains(link.URL, "youtube.com") || strings.Contains(link.URL, "youtu.be") {
			linkText = "VIDEO"
		}
		builder.WriteString(fmt.Sprintf("%s [%s](%s)\n\n", link.MyTake, linkText, link.URL))
	}

	// Add "What to Print" section if present
	if len(content.PrintItems) > 0 {
		builder.WriteString("## What to Print This Week\n\n")
		builder.WriteString("This newsletter started out on 3D printing. If you haven't had any contact with it, you should, it's great! Here's the most interesting and fun projects I saw last week.\n\n")

		for _, item := range content.PrintItems {
			builder.WriteString(fmt.Sprintf("### %s\n\n", item.Title))

			if item.ImageURL != "" {
				builder.WriteString(fmt.Sprintf("![%s](%s)\n\n", item.Title, item.ImageURL))
			}

			if item.Description != "" {
				builder.WriteString(fmt.Sprintf("%s\n\n", item.Description))
			}

			if item.LinkURL != "" {
				builder.WriteString(fmt.Sprintf("[visit model page](%s)\n\n", item.LinkURL))
			}

			builder.WriteString("---\n\n")
		}
	}

	// Add footer section if present
	if content.FooterContent != "" {
		builder.WriteString(content.FooterContent)
		builder.WriteString("\n")
	}

	return builder.String()
}

func formatNewsletterContentHTML(content *models.Content) string {
	var builder strings.Builder

	// Start with complete HTML document structure with max-width container
	builder.WriteString(`<!DOCTYPE html>`)
	builder.WriteString(`<html>`)
	builder.WriteString(`<head>`)
	builder.WriteString(`<meta charset="utf-8">`)
	builder.WriteString(`<style>`)
	builder.WriteString(`@import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@600&display=swap');`)
	builder.WriteString(`h1, h2, h3 { font-family: 'IBM Plex Mono', monospace !important; font-weight: 600 !important; }`)
	builder.WriteString(`.container { max-width: 600px; margin: 0 auto; padding: 20px; }`)
	builder.WriteString(`</style>`)
	builder.WriteString(`</head>`)
	builder.WriteString(`<body style="margin: 0; padding: 0;">`)

	// Main container with max width
	builder.WriteString(`<div class="container" style="max-width: 600px; margin: 0 auto; padding: 20px;">`)

	// Top banner image - full width within container
	builder.WriteString(`<img src="https://liquid.engineer/images/banner.png" style="width: 100%; height: auto; display: block; margin-bottom: 24px;" alt="The Liquid Engineer">`)

	// Header without green border
	builder.WriteString(`<div style="margin: 24px 0;">`)
	builder.WriteString(fmt.Sprintf(`<h1 style="font-family: 'IBM Plex Mono', monospace; font-size: 36px; color: #12363f; font-weight: 600; line-height: 1.5; margin: 0;">%s</h1>`, content.Metadata.Title))
	builder.WriteString(fmt.Sprintf(`<p style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 18px; color: #4e585a; font-weight: 400; line-height: 1.5; margin-top: 8px;">The Liquid Engineer – Issue No. %s</p>`, strings.TrimPrefix(content.Metadata.ContentID, "issue")))
	builder.WriteString(`</div>`)

	// Process the thought piece to use absolute URLs for images
	processedThoughtPiece := processImagePathsForNewsletter(content.ThoughtPiece, content.Metadata.ContentID)

	// Convert thought piece from markdown to HTML using goldmark
	thoughtPieceHTML := convertMarkdownToBasicHTML(processedThoughtPiece)
	builder.WriteString(thoughtPieceHTML)

	// Links section
	builder.WriteString(`<h2 style="font-family: 'IBM Plex Mono', monospace; font-size: 36px; color: #11363F; font-weight: 600; line-height: 1.5; margin-top: 32px;">What I Learned This Week</h2>`)

	for _, link := range content.Links {
		linkText := "LINK"
		if strings.Contains(link.URL, "youtube.com") || strings.Contains(link.URL, "youtu.be") {
			linkText = "VIDEO"
		}
		builder.WriteString(fmt.Sprintf(`<p style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 18px; color: #353535; font-weight: 400; line-height: 1.5;">%s <a href="%s" style="color: #0066cc; text-decoration: underline;">%s</a></p>`, link.MyTake, link.URL, linkText))
	}

	// Print items section if present
	if len(content.PrintItems) > 0 {
		builder.WriteString(`<h2 style="font-family: 'IBM Plex Mono', monospace; font-size: 36px; color: #11363F; font-weight: 600; line-height: 1.5; margin-top: 32px;">What to Print This Week</h2>`)
		builder.WriteString(`<p style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 18px; color: #353535; font-weight: 400; line-height: 1.5;">This newsletter started out on 3D printing. If you haven't had any contact with it, you should, it's great! Here's the most interesting and fun projects I saw last week.</p>`)

		for _, item := range content.PrintItems {
			// Create a card layout with table structure
			builder.WriteString(`<table cellpadding="0" cellspacing="0" style="width: 100%; margin: 24px 0; border-collapse: collapse;">`)
			builder.WriteString(`<tr>`)

			// Image column (150px fixed width) - make image clickable
			builder.WriteString(`<td style="width: 150px; padding-right: 20px; vertical-align: top;">`)
			if item.ImageURL != "" {
				builder.WriteString(fmt.Sprintf(`<a href="%s" style="display: block; text-decoration: none;">`, item.LinkURL))
				builder.WriteString(fmt.Sprintf(`<img src="%s" style="width: 150px; height: 150px; object-fit: cover; border-radius: 8px; display: block;">`, item.ImageURL))
				builder.WriteString(`</a>`)
			} else {
				builder.WriteString(`<div style="width: 150px; height: 150px; background-color: #f0f0f0; border-radius: 8px;"></div>`)
			}
			builder.WriteString(`</td>`)

			// Content column
			builder.WriteString(`<td style="vertical-align: top;">`)
			builder.WriteString(fmt.Sprintf(`<h3 style="font-family: 'IBM Plex Mono', monospace; font-size: 24px; color: #11363F; font-weight: 600; line-height: 1.3; margin: 0 0 8px 0;">%s</h3>`, item.Title))

			if item.Description != "" {
				builder.WriteString(fmt.Sprintf(`<p style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 16px; color: #353535; font-weight: 400; line-height: 1.5; margin: 0 0 12px 0;">%s</p>`, item.Description))
			}

			// Make button clickable with proper <a> tag
			builder.WriteString(fmt.Sprintf(`<a href="%s" style="text-decoration: none; display: inline-block;">`, item.LinkURL))
			builder.WriteString(`<span style="display: inline-block; background-color: #eab2bb; color: #ffffff; border-radius: 4px; font-family: -apple-system, BlinkMacSystemFont, sans-serif; padding: 10px 16px; font-size: 14px; font-weight: 700;">visit model page</span>`)
			builder.WriteString(`</a>`)
			builder.WriteString(`</td>`)

			builder.WriteString(`</tr>`)
			builder.WriteString(`</table>`)
		}
	}

	// Footer if present
	if content.FooterContent != "" {
		builder.WriteString(`<div style="background-color: #f9f9f9; padding: 24px; margin: 24px 0; border-radius: 8px;">`)
		builder.WriteString(`<table cellpadding="0" cellspacing="0" style="width: 100%; border-collapse: collapse;">`)
		builder.WriteString(`<tr>`)

		// Content column on the left
		builder.WriteString(`<td style="vertical-align: middle; padding-right: 20px;">`)
		footerLines := strings.Split(content.FooterContent, "\n")
		for _, line := range footerLines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "##") {
				title := strings.TrimSpace(strings.TrimPrefix(line, "##"))
				builder.WriteString(fmt.Sprintf(`<h2 style="font-family: 'IBM Plex Mono', monospace; font-size: 32px; color: #000000; font-weight: 600; line-height: 1.5; margin: 0 0 12px 0;">%s</h2>`, title))
			} else if line != "" {
				line = convertMarkdownLinksToHTML(line)
				builder.WriteString(fmt.Sprintf(`<p style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 16px; color: #4d4d4d; font-weight: 400; line-height: 1.5; margin: 8px 0;">%s</p>`, line))
			}
		}
		builder.WriteString(`</td>`)

		// Avatar column on the right (180px round image)
		builder.WriteString(`<td style="width: 180px; vertical-align: middle; text-align: right;">`)
		builder.WriteString(`<img src="https://liquid.engineer/images/avatar.jpg" style="width: 180px; height: 180px; border-radius: 50%; object-fit: cover; display: block;" alt="Stefan Munz">`)
		builder.WriteString(`</td>`)

		builder.WriteString(`</tr>`)
		builder.WriteString(`</table>`)
		builder.WriteString(`</div>`)
	}

	// Close container div
	builder.WriteString(`</div>`)

	builder.WriteString(`</body>`)
	builder.WriteString(`</html>`)

	return builder.String()
}

func convertMarkdownToBasicHTML(markdown string) string {
	// Step 1: Convert markdown to HTML with goldmark
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &buf); err != nil {
		log.Printf("ERROR: Failed to convert markdown: %v", err)
		return markdown
	}

	// Step 2: Wrap in full HTML document with CSS styles
	fullHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
<style type="text/css">
	p {
		font-family: -apple-system, BlinkMacSystemFont, sans-serif;
		font-size: 18px;
		color: #353535;
		font-weight: 400;
		line-height: 28px;
		margin: 0 0 16px 0;
	}
	ul {
		font-family: -apple-system, BlinkMacSystemFont, sans-serif;
		font-size: 18px;
		color: #353535;
		line-height: 28px;
		margin: 16px 0;
		padding-left: 24px;
	}
	li {
		margin-bottom: 8px;
	}
	h2 {
		font-family: 'IBM Plex Mono', monospace;
		font-size: 24px;
		color: #11363F;
		font-weight: 600;
		line-height: 1.5;
		margin-top: 32px;
	}
	h3 {
		font-family: 'IBM Plex Mono', monospace;
		font-size: 20px;
		color: #11363F;
		font-weight: 600;
		line-height: 1.5;
		margin-top: 24px;
	}
	a {
		color: #0066cc;
		text-decoration: underline;
	}
	img {
		width: 100%%;
		max-width: 600px;
		height: auto;
		display: block;
		margin: 24px auto;
	}
</style>
</head>
<body>%s</body>
</html>`, buf.String())

	// Step 3: Use premailer to inline CSS
	prem, err := premailer.NewPremailerFromString(fullHTML, premailer.NewOptions())
	if err != nil {
		log.Printf("ERROR: Failed to create premailer: %v", err)
		return buf.String()
	}

	inlinedHTML, err := prem.Transform()
	if err != nil {
		log.Printf("ERROR: Failed to transform with premailer: %v", err)
		return buf.String()
	}

	// Step 4: Extract just the body content (strip <html><head><body> wrapper)
	bodyStart := strings.Index(inlinedHTML, "<body>")
	bodyEnd := strings.Index(inlinedHTML, "</body>")
	if bodyStart == -1 || bodyEnd == -1 {
		log.Printf("WARNING: Could not extract body content, returning full HTML")
		return inlinedHTML
	}

	bodyContent := inlinedHTML[bodyStart+6 : bodyEnd]
	return bodyContent
}

// Helper function to convert markdown images to HTML
func convertMarkdownImagesToHTML(text string) string {
	// Use regex to find markdown images ![alt](url)
	re := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	result := re.ReplaceAllStringFunc(text, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		altText := submatches[1]
		imageURL := submatches[2]

		// Return HTML img tag with styling
		return fmt.Sprintf(`<img src="%s" alt="%s" style="width: 100%%; max-width: 600px; height: auto; display: block; margin: 24px auto;">`, imageURL, altText)
	})

	return result
}

// Helper function to convert markdown links to HTML
func convertMarkdownLinksToHTML(text string) string {
	// Simple regex replacement for [text](url) pattern
	// This is a basic implementation - could be improved with proper regex
	for {
		start := strings.Index(text, "[")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], "](")
		if end == -1 {
			break
		}
		urlEnd := strings.Index(text[start+end:], ")")
		if urlEnd == -1 {
			break
		}

		linkText := text[start+1 : start+end]
		url := text[start+end+2 : start+end+urlEnd]
		htmlLink := fmt.Sprintf(`<a href="%s" style="color: #0000ff;">%s</a>`, url, linkText)

		text = text[:start] + htmlLink + text[start+end+urlEnd+1:]
	}
	return text
}
