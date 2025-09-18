package handlers

import (
	"distribute/config"
	"distribute/models"
	"distribute/utils"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type BufferPost struct {
	Text      string
	ProfileID string
	Platform  models.Platform
	IsReply   bool
	ReplyToID string
}

func HandleBufferScheduling(cfg *config.Config, content *models.Content, dryRun bool) error {
	// Generate blog post URL
	blogURL := utils.BuildBlogURL(cfg.BlogBaseURL, content.OriginalPath)
	
	// Collect all posts to be created
	var allPosts []BufferPost
	
	// Only generate posts for the first profile to avoid duplicates
	if len(cfg.BufferProfileIDs) > 0 {
		profileID := cfg.BufferProfileIDs[0]
		platform, exists := cfg.ProfilePlatformMap[profileID]
		if !exists {
			log.Printf("WARNING: Unknown platform for profile ID %s, using Twitter limits", profileID)
			platform = models.PlatformTwitter
		}
		
		// Generate posts for just the first platform
		platformPosts := generatePlatformPosts(content, blogURL, platform, profileID)
		allPosts = append(allPosts, platformPosts...)
	}
	
	// MODIFIED: Always display manual posting output instead of API calls
	// When Buffer's new API is available, we can re-enable the API calls below
	displayManualPostingOutput(allPosts, cfg)
	
	// COMMENTED OUT: Original Buffer API code - will re-enable when new API is available
	/*
	if dryRun {
		displayDryRunOutput(allPosts, cfg)
		return nil
	}
	
	// Schedule posts with Buffer
	for _, post := range allPosts {
		var err error
		if post.IsReply {
			err = scheduleBufferReply(cfg.BufferAccessToken, post)
		} else {
			err = scheduleBufferPost(cfg.BufferAccessToken, post)
		}
		
		if err != nil {
			log.Printf("ERROR: Failed to schedule post to Buffer: %v", err)
			// Continue with other posts
		} else {
			log.Printf("INFO: Successfully scheduled post to Buffer for %s", post.Platform)
		}
	}
	*/
	
	return nil
}

func generatePlatformPosts(content *models.Content, blogURL string, platform models.Platform, profileID string) []BufferPost {
	var posts []BufferPost

	// Strip markdown links from the thought piece for social media
	strippedThoughtPiece := utils.StripMarkdownLinks(content.ThoughtPiece)

	// 1. Main thought piece post - expanded to ~800 characters to leave room for links
	thoughtPieceText := fmt.Sprintf("%s\n\nRead the full article with links here: %s?utm_source=social&utm_medium=social\n\nSubscribe to my newsletter to get pieces like this into your inbox automatically, every week! Plus the most interesting links I found this week. https://liquid.engineer/\n\n%s",
		utils.TruncateText(strippedThoughtPiece, 800), // Reduced to 800 to accommodate additional text
		blogURL,
		content.Metadata.SocialMediaHashtags,
	)
	
	// No threading - just single posts
	posts = append(posts, BufferPost{
		Text:      thoughtPieceText,
		ProfileID: profileID,
		Platform:  platform,
	})
	
	// 2. Curated links posts
	for _, link := range content.Links {
		// Simple format: MyTake + URL (no quotes, no hashtags, no extra text)
		linkText := fmt.Sprintf("%s %s", link.MyTake, link.URL)
		
		// No threading - just single posts
		posts = append(posts, BufferPost{
			Text:      linkText,
			ProfileID: profileID,
			Platform:  platform,
		})
	}
	
	return posts
}

func scheduleBufferPost(accessToken string, post BufferPost) error {
	endpoint := "https://api.bufferapp.com/1/updates/create.json"
	
	data := url.Values{}
	data.Set("text", post.Text)
	data.Add("profile_ids[]", post.ProfileID)
	
	// Only the first thought piece post gets "top" priority
	if !post.IsReply && strings.Contains(post.Text, "Read the full article with links") {
		data.Set("top", "true")
	}
	
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Buffer API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func scheduleBufferReply(accessToken string, post BufferPost) error {
	// For threading, we'll use the same endpoint but with reply parameters
	// Note: Buffer's threading support varies by platform
	// This is a simplified implementation
	return scheduleBufferPost(accessToken, post)
}

func displayDryRunOutput(posts []BufferPost, cfg *config.Config) {
	fmt.Println("=== DRY RUN MODE ===")
	
	currentProfile := ""
	postCounter := 1
	
	for _, post := range posts {
		if post.ProfileID != currentProfile {
			currentProfile = post.ProfileID
			postCounter = 1
			
			platformName := "Unknown"
			if platform, exists := cfg.ProfilePlatformMap[post.ProfileID]; exists {
				platformName = string(platform)
			}
			
			fmt.Printf("📱 BUFFER - %s Profile (ID: %s):\n", strings.Title(platformName), post.ProfileID)
		}
		
		postType := "Link"
		if strings.Contains(post.Text, "Read the full article") {
			postType = "Main - top of queue"
		}
		if post.IsReply {
			postType = "Reply/Thread continuation"
		}
		
		fmt.Printf("Post %d [%s]:\n", postCounter, postType)
		fmt.Printf("\"%s\"\n\n", post.Text)
		postCounter++
	}
	
	fmt.Println("=== END DRY RUN ===")
}

func displayManualPostingOutput(posts []BufferPost, cfg *config.Config) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SOCIAL MEDIA POSTS - READY FOR BUFFER")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	
	linkCounter := 1
	for _, post := range posts {
		if strings.Contains(post.Text, "Read the full article with links") {
			fmt.Printf("\n[MAIN POST - Newsletter Summary]:\n")
		} else {
			fmt.Printf("\n[CURATED LINK %d]:\n", linkCounter)
			linkCounter++
		}
		
		// Display the actual post content in a copy-friendly format
		fmt.Println(strings.Repeat("-", 70))
		fmt.Println(post.Text)
		fmt.Println(strings.Repeat("-", 70))
	}
	
	fmt.Printf("\nTotal posts: %d\n", len(posts))
	
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("INSTRUCTIONS:")
	fmt.Println("1. Copy each post above and paste into Buffer")
	fmt.Println("2. Edit the main post to your desired length (currently ~1000 chars)")
	fmt.Println("3. Schedule main post at top of queue, link posts in regular queue")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))
}