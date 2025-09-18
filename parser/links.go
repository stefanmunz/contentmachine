package parser

import (
	"distribute/models"
	"fmt"
	"strings"
)

func ParseLinks(linksSection string) ([]models.CuratedLink, error) {
	var links []models.CuratedLink

	// Split by horizontal rules
	linkBlocks := strings.Split(linksSection, "\n---\n")

	for _, block := range linkBlocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		link, err := parseSingleLink(block)
		if err != nil {
			return nil, fmt.Errorf("failed to parse link block: %w\nBlock content:\n%s", err, block)
		}
		links = append(links, link)
	}

	return links, nil
}

func parseSingleLink(block string) (models.CuratedLink, error) {
	var link models.CuratedLink

	lines := strings.Split(block, "\n")

	// Look for the title line
	titleFound := false
	urlFound := false
	myTakeFound := false
	keywordFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse Title
		if strings.HasPrefix(line, "- **Title:**") {
			titleText := strings.TrimPrefix(line, "- **Title:**")
			link.Title = strings.TrimSpace(titleText)
			titleFound = true
		}

		// Parse URL
		if strings.HasPrefix(line, "- **URL:**") {
			urlText := strings.TrimPrefix(line, "- **URL:**")
			link.URL = strings.TrimSpace(urlText)
			urlFound = true
		}

		// Parse MyTake
		if strings.HasPrefix(line, "- **MyTake:**") {
			myTakeText := strings.TrimPrefix(line, "- **MyTake:**")
			link.MyTake = strings.TrimSpace(myTakeText)
			myTakeFound = true
		}

		// Parse Keyword
		if strings.HasPrefix(line, "- **Keyword:**") {
			keywordText := strings.TrimPrefix(line, "- **Keyword:**")
			link.Keyword = strings.TrimSpace(keywordText)
			keywordFound = true
		}
	}

	// Title is now optional - it's integrated into MyTake
	if !titleFound {
		link.Title = "" // Empty title is fine
	}
	if !urlFound {
		return link, fmt.Errorf("URL not found in link block")
	}
	if !myTakeFound {
		return link, fmt.Errorf("MyTake not found in link block")
	}

	// Keyword is optional - default to "link" if not specified
	if !keywordFound {
		link.Keyword = "link"
	}

	return link, nil
}
