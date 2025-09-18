package parser

import (
	"distribute/models"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	linksSeparator  = "<!--LINKS_SEPARATOR-->"
	printSeparator  = "<!--PRINT_SEPARATOR-->"
	footerSeparator = "<!--FOOTER_SEPARATOR-->"
)

func ParseMarkdownFile(filePath string) (*models.Content, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	content := &models.Content{
		OriginalPath: filePath,
	}

	// Split frontmatter and body
	parts := strings.SplitN(string(fileContent), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown format: frontmatter not found")
	}

	// Parse frontmatter
	if err := yaml.Unmarshal([]byte(parts[1]), &content.Metadata); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Split body into thought piece and links
	bodyParts := strings.SplitN(parts[2], linksSeparator, 2)
	if len(bodyParts) < 2 {
		return nil, fmt.Errorf("invalid markdown format: %s not found", linksSeparator)
	}

	content.ThoughtPiece = strings.TrimSpace(bodyParts[0])

	// Parse the rest of the content
	remainingContent := strings.TrimSpace(bodyParts[1])

	// Check for print section
	if strings.Contains(remainingContent, printSeparator) {
		parts := strings.SplitN(remainingContent, printSeparator, 2)
		linksSection := strings.TrimSpace(parts[0])
		remainingContent = strings.TrimSpace(parts[1])

		// Parse links
		content.Links, err = ParseLinks(linksSection)
		if err != nil {
			return nil, fmt.Errorf("failed to parse links: %w", err)
		}

		// Check for footer section
		if strings.Contains(remainingContent, footerSeparator) {
			parts := strings.SplitN(remainingContent, footerSeparator, 2)
			printSection := strings.TrimSpace(parts[0])
			content.FooterContent = strings.TrimSpace(parts[1])

			// Parse print items
			content.PrintItems = ParsePrintItems(printSection)
		} else {
			// Only print section, no footer
			content.PrintItems = ParsePrintItems(remainingContent)
		}
	} else if strings.Contains(remainingContent, footerSeparator) {
		// No print section, but has footer
		parts := strings.SplitN(remainingContent, footerSeparator, 2)
		linksSection := strings.TrimSpace(parts[0])
		content.FooterContent = strings.TrimSpace(parts[1])

		// Parse links
		content.Links, err = ParseLinks(linksSection)
		if err != nil {
			return nil, fmt.Errorf("failed to parse links: %w", err)
		}
	} else {
		// Only links section
		content.Links, err = ParseLinks(remainingContent)
		if err != nil {
			return nil, fmt.Errorf("failed to parse links: %w", err)
		}
	}

	return content, nil
}

// ParsePrintItems parses the "What to Print" section
func ParsePrintItems(content string) []models.PrintItem {
	var items []models.PrintItem

	// Split by ### headers
	sections := strings.Split(content, "###")

	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" || strings.HasPrefix(section, "#") {
			continue
		}

		item := models.PrintItem{}

		// Get title (first line)
		lines := strings.Split(section, "\n")
		if len(lines) > 0 {
			item.Title = strings.TrimSpace(lines[0])
		}

		// Parse the rest for description, image URL, and link
		for _, line := range lines[1:] {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "visit model page") && strings.Contains(line, "(") {
				// Extract link URL
				start := strings.Index(line, "(")
				end := strings.Index(line, ")")
				if start > 0 && end > start {
					item.LinkURL = line[start+1 : end]
				}
			} else if strings.HasPrefix(line, "![") {
				// Markdown image syntax
				start := strings.Index(line, "](")
				end := strings.LastIndex(line, ")")
				if start > 0 && end > start {
					item.ImageURL = line[start+2 : end]
				}
			} else if line != "" && !strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "*") && !strings.HasPrefix(line, "---") {
				// Regular description text
				if item.Description == "" {
					item.Description = line
				} else {
					item.Description += " " + line
				}
			}
		}

		if item.Title != "" {
			items = append(items, item)
		}
	}

	return items
}
