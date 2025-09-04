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
	"strings"
)

// Kit v4 API request structures
type KitBroadcastRequest struct {
	EmailTemplateID int    `json:"email_template_id"`
	Subject         string `json:"subject"`
	Content         string `json:"content"`
	Public          bool   `json:"public"`
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
		
		fmt.Printf("\nHTML Preview (first 1500 chars):\n")
		preview := richHTML
		if len(preview) > 1500 {
			preview = preview[:1500] + "...[truncated]"
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

func formatNewsletterContentMarkdown(content *models.Content) string {
	var builder strings.Builder
	
	// Add header with title and issue number
	builder.WriteString(fmt.Sprintf("# %s\n\n", content.Metadata.Title))
	builder.WriteString(fmt.Sprintf("*The Liquid Engineer – Issue No. %s*\n\n", strings.TrimPrefix(content.Metadata.ContentID, "issue")))
	builder.WriteString("---\n\n")
	
	// Add the thought piece
	builder.WriteString(content.ThoughtPiece)
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
	builder.WriteString(`@import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:ital@1&display=swap');`)
	builder.WriteString(`h1, h2, h3 { font-family: 'IBM Plex Mono', monospace !important; font-style: italic !important; }`)
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
	builder.WriteString(fmt.Sprintf(`<h1 style="font-family: 'IBM Plex Mono', monospace; font-style: italic; font-size: 36px; color: #12363f; font-weight: 400; line-height: 1.5; margin: 0;">%s</h1>`, content.Metadata.Title))
	builder.WriteString(fmt.Sprintf(`<p style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 18px; color: #4e585a; font-weight: 400; line-height: 1.5; margin-top: 8px;">The Liquid Engineer – Issue No. %s</p>`, strings.TrimPrefix(content.Metadata.ContentID, "issue")))
	builder.WriteString(`</div>`)
	
	// Article banner image if it exists for this issue
	issueFolder := strings.TrimPrefix(content.Metadata.ContentID, "issue")
	if issueFolder != "" {
		builder.WriteString(fmt.Sprintf(`<img src="https://liquid.engineer/issues/%s/banner.jpg" style="width: 100%%; height: auto; display: block; margin: 24px 0;" alt="Issue %s Banner">`, issueFolder, issueFolder))
	}
	
	// Thought piece
	paragraphs := strings.Split(content.ThoughtPiece, "\n\n")
	for _, para := range paragraphs {
		if strings.TrimSpace(para) != "" {
			para = convertMarkdownLinksToHTML(para)
			builder.WriteString(fmt.Sprintf(`<p style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 18px; color: #353535; font-weight: 400; line-height: 1.5;">%s</p>`, para))
		}
	}
	
	// Links section
	builder.WriteString(`<h2 style="font-family: 'IBM Plex Mono', monospace; font-style: italic; font-size: 32px; color: #11363F; font-weight: 400; line-height: 1.5; margin-top: 32px;">What I Learned This Week</h2>`)
	
	for _, link := range content.Links {
		linkText := "LINK"
		if strings.Contains(link.URL, "youtube.com") || strings.Contains(link.URL, "youtu.be") {
			linkText = "VIDEO"
		}
		builder.WriteString(fmt.Sprintf(`<p style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 18px; color: #353535; font-weight: 400; line-height: 1.5;">%s <a href="%s" style="color: #0066cc; text-decoration: underline;">%s</a></p>`, link.MyTake, link.URL, linkText))
	}
	
	// Print items section if present
	if len(content.PrintItems) > 0 {
		builder.WriteString(`<h2 style="font-family: 'IBM Plex Mono', monospace; font-style: italic; font-size: 32px; color: #11363F; font-weight: 400; line-height: 1.5; margin-top: 32px;">What to Print This Week</h2>`)
		builder.WriteString(`<p style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 18px; color: #353535; font-weight: 400; line-height: 1.5;">This newsletter started out on 3D printing. If you haven't had any contact with it, you should, it's great! Here's the most interesting and fun projects I saw last week.</p>`)
		
		for _, item := range content.PrintItems {
			// Create a clickable card layout with table structure
			builder.WriteString(fmt.Sprintf(`<a href="%s" style="text-decoration: none; color: inherit; display: block;">`, item.LinkURL))
			builder.WriteString(`<table cellpadding="0" cellspacing="0" style="width: 100%; margin: 24px 0; border-collapse: collapse;">`)
			builder.WriteString(`<tr>`)
			
			// Image column (150px fixed width)
			builder.WriteString(`<td style="width: 150px; padding-right: 20px; vertical-align: top;">`)
			if item.ImageURL != "" {
				builder.WriteString(fmt.Sprintf(`<img src="%s" style="width: 150px; height: 150px; object-fit: cover; border-radius: 8px; display: block;">`, item.ImageURL))
			} else {
				builder.WriteString(`<div style="width: 150px; height: 150px; background-color: #f0f0f0; border-radius: 8px;"></div>`)
			}
			builder.WriteString(`</td>`)
			
			// Content column
			builder.WriteString(`<td style="vertical-align: top;">`)
			builder.WriteString(fmt.Sprintf(`<h3 style="font-family: 'IBM Plex Mono', monospace; font-style: italic; font-size: 24px; color: #11363F; font-weight: 400; line-height: 1.3; margin: 0 0 8px 0;">%s</h3>`, item.Title))
			
			if item.Description != "" {
				builder.WriteString(fmt.Sprintf(`<p style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 16px; color: #353535; font-weight: 400; line-height: 1.5; margin: 0 0 12px 0;">%s</p>`, item.Description))
			}
			
			builder.WriteString(`<span style="display: inline-block; background-color: #eab2bb; color: #ffffff; border-radius: 4px; font-family: -apple-system, BlinkMacSystemFont, sans-serif; padding: 10px 16px; font-size: 14px; font-weight: 700;">visit model page</span>`)
			builder.WriteString(`</td>`)
			
			builder.WriteString(`</tr>`)
			builder.WriteString(`</table>`)
			builder.WriteString(`</a>`)
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
				builder.WriteString(fmt.Sprintf(`<h2 style="font-family: 'IBM Plex Mono', monospace; font-style: italic; font-size: 32px; color: #000000; font-weight: 500; line-height: 1.5; margin: 0 0 12px 0;">%s</h2>`, title))
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
	var html strings.Builder
	lines := strings.Split(markdown, "\n")
	inParagraph := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Handle headers
		if strings.HasPrefix(trimmed, "### ") {
			if inParagraph {
				html.WriteString("</p>\n")
				inParagraph = false
			}
			html.WriteString(fmt.Sprintf("<h3>%s</h3>\n", strings.TrimPrefix(trimmed, "### ")))
		} else if strings.HasPrefix(trimmed, "## ") {
			if inParagraph {
				html.WriteString("</p>\n")
				inParagraph = false
			}
			html.WriteString(fmt.Sprintf("<h2>%s</h2>\n", strings.TrimPrefix(trimmed, "## ")))
		} else if strings.HasPrefix(trimmed, "# ") {
			if inParagraph {
				html.WriteString("</p>\n")
				inParagraph = false
			}
			html.WriteString(fmt.Sprintf("<h1>%s</h1>\n", strings.TrimPrefix(trimmed, "# ")))
		} else if trimmed == "---" {
			if inParagraph {
				html.WriteString("</p>\n")
				inParagraph = false
			}
			html.WriteString("<hr>\n")
		} else if trimmed == "" {
			// Empty line - close paragraph if open
			if inParagraph {
				html.WriteString("</p>\n")
				inParagraph = false
			}
		} else {
			// Regular text - convert markdown elements
			processed := trimmed
			
			// Convert bold
			processed = strings.ReplaceAll(processed, "**", "")
			
			// Convert italic
			if strings.HasPrefix(processed, "*") && strings.HasSuffix(processed, "*") && len(processed) > 2 {
				processed = fmt.Sprintf("<em>%s</em>", processed[1:len(processed)-1])
			}
			
			// Convert markdown links to HTML
			processed = convertMarkdownLinksToHTML(processed)
			
			// Convert markdown images to HTML
			if strings.HasPrefix(processed, "![") {
				start := strings.Index(processed, "](")
				end := strings.LastIndex(processed, ")")
				if start > 0 && end > start {
					alt := processed[2:start]
					src := processed[start+2:end]
					processed = fmt.Sprintf(`<img src="%s" alt="%s" style="max-width: 100%%;">`, src, alt)
				}
			}
			
			// Start or continue paragraph
			if !inParagraph {
				html.WriteString("<p>")
				inParagraph = true
			} else if i > 0 && lines[i-1] != "" {
				html.WriteString(" ")
			}
			html.WriteString(processed)
		}
	}
	
	// Close any open paragraph
	if inParagraph {
		html.WriteString("</p>\n")
	}
	
	return html.String()
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