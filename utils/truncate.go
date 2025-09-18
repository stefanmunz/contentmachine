package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// TruncateText truncates text to maxLength characters, breaking at word boundaries and adding ellipsis
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	
	// Find the last space before maxLength
	truncated := text[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")
	
	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}
	
	return truncated + "..."
}

// CreateThreadedPosts splits long text into multiple posts for threading
func CreateThreadedPosts(text string, charLimit int, hashtags string) []string {
	var posts []string
	
	// Reserve space for thread numbering (e.g., " (1/3)")
	threadIndicatorSpace := 10
	effectiveLimit := charLimit - threadIndicatorSpace
	
	// Split text into words
	words := strings.Fields(text)
	currentPost := ""
	
	for _, word := range words {
		testPost := currentPost
		if testPost != "" {
			testPost += " "
		}
		testPost += word
		
		// Check if adding this word would exceed the limit
		if len(testPost) > effectiveLimit {
			if currentPost != "" {
				posts = append(posts, currentPost)
				currentPost = word
			} else {
				// Single word is too long, truncate it
				posts = append(posts, TruncateText(word, effectiveLimit))
				currentPost = ""
			}
		} else {
			currentPost = testPost
		}
	}
	
	// Add remaining text
	if currentPost != "" {
		posts = append(posts, currentPost)
	}
	
	// Add thread numbering
	if len(posts) > 1 {
		for i := range posts {
			posts[i] = fmt.Sprintf("%s (%d/%d)", posts[i], i+1, len(posts))
		}
	}
	
	return posts
}

// StripMarkdownLinks removes markdown links from text, keeping only the link text
// For example: [link text](url) becomes "link text"
func StripMarkdownLinks(text string) string {
	// Pattern to match markdown links [text](url)
	re := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	// Replace with just the link text (captured group 1)
	return re.ReplaceAllString(text, "$1")
}