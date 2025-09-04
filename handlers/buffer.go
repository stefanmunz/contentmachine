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
	
	// Process for each profile
	for _, profileID := range cfg.BufferProfileIDs {
		platform, exists := cfg.ProfilePlatformMap[profileID]
		if !exists {
			log.Printf("WARNING: Unknown platform for profile ID %s, using Twitter limits", profileID)
			platform = models.PlatformTwitter
		}
		
		// Generate posts for this platform
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
	charLimit := models.GetCharLimit(platform)
	
	// 1. Main thought piece post(s)
	thoughtPieceText := fmt.Sprintf("%s... Read the full article: %s?utm_source=social&utm_medium=social %s",
		utils.TruncateText(content.ThoughtPiece, 200),
		blogURL,
		content.Metadata.SocialMediaHashtags,
	)
	
	// Check if threading is needed for thought piece
	if len(thoughtPieceText) > charLimit {
		threadPosts := utils.CreateThreadedPosts(thoughtPieceText, charLimit, content.Metadata.SocialMediaHashtags)
		for i, text := range threadPosts {
			posts = append(posts, BufferPost{
				Text:      text,
				ProfileID: profileID,
				Platform:  platform,
				IsReply:   i > 0,
			})
		}
	} else {
		posts = append(posts, BufferPost{
			Text:      thoughtPieceText,
			ProfileID: profileID,
			Platform:  platform,
		})
	}
	
	// 2. Curated links posts
	for _, link := range content.Links {
		linkText := fmt.Sprintf(`"%s"

A great read on "%s" from my weekly newsletter.
%s
%s`, link.MyTake, link.Title, link.URL, content.Metadata.SocialMediaHashtags)
		
		// Check if threading is needed for this link
		if len(linkText) > charLimit {
			threadPosts := utils.CreateThreadedPosts(linkText, charLimit, content.Metadata.SocialMediaHashtags)
			for i, text := range threadPosts {
				posts = append(posts, BufferPost{
					Text:      text,
					ProfileID: profileID,
					Platform:  platform,
					IsReply:   i > 0,
				})
			}
		} else {
			posts = append(posts, BufferPost{
				Text:      linkText,
				ProfileID: profileID,
				Platform:  platform,
			})
		}
	}
	
	return posts
}

func scheduleBufferPost(accessToken string, post BufferPost) error {
	endpoint := "https://api.bufferapp.com/1/updates/create.json"
	
	data := url.Values{}
	data.Set("text", post.Text)
	data.Add("profile_ids[]", post.ProfileID)
	
	// Only the first thought piece post gets "top" priority
	if !post.IsReply && strings.Contains(post.Text, "Read the full article") {
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
	fmt.Println("=== DRY RUN MODE ===\n")
	
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
	fmt.Println("📱 SOCIAL MEDIA POSTS - COPY & PASTE TO BUFFER")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	
	// Group posts by platform
	platformPosts := make(map[models.Platform][]BufferPost)
	for _, post := range posts {
		platformPosts[post.Platform] = append(platformPosts[post.Platform], post)
	}
	
	// Display posts for each platform
	for _, platform := range []models.Platform{models.PlatformTwitter, models.PlatformLinkedIn, models.PlatformBluesky} {
		posts, exists := platformPosts[platform]
		if !exists || len(posts) == 0 {
			continue
		}
		
		fmt.Printf("\n%s %s POSTS %s\n", strings.Repeat("-", 30), strings.ToUpper(string(platform)), strings.Repeat("-", 30))
		
		isThread := false
		threadNum := 1
		for i, post := range posts {
			// Detect thread starts and continuations
			if post.IsReply {
				if !isThread {
					isThread = true
					threadNum = 2
				} else {
					threadNum++
				}
				fmt.Printf("\n🔗 Thread Part %d (post as reply to previous):\n", threadNum)
			} else {
				if isThread {
					// Reset thread tracking
					isThread = false
					threadNum = 1
				}
				
				if strings.Contains(post.Text, "Read the full article") {
					fmt.Printf("\n📌 MAIN POST (add to top of queue):\n")
					if i+1 < len(posts) && posts[i+1].IsReply {
						fmt.Printf("⚠️  This will be a THREAD - post the following parts as replies\n")
					}
				} else {
					fmt.Printf("\n📎 Link Post %d:\n", i+1-threadNum+1)
				}
			}
			
			// Display the actual post content in a copy-friendly format
			fmt.Println(strings.Repeat("-", 70))
			fmt.Println(post.Text)
			fmt.Println(strings.Repeat("-", 70))
		}
		
		fmt.Printf("\n✅ Total posts for %s: %d\n", platform, len(posts))
		
		// Platform-specific instructions
		switch platform {
		case models.PlatformTwitter:
			fmt.Println("💡 Twitter/X: Posts longer than 280 chars are split into threads")
		case models.PlatformLinkedIn:
			fmt.Println("💡 LinkedIn: 3000 char limit, no threading needed")
		case models.PlatformBluesky:
			fmt.Println("💡 Bluesky: Posts longer than 300 chars are split into threads")
		}
	}
	
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("📝 INSTRUCTIONS:")
	fmt.Println("1. Copy each post above and paste into Buffer")
	fmt.Println("2. For MAIN POSTS: Add to top of queue")
	fmt.Println("3. For THREADS: Post subsequent parts as replies to the previous post")
	fmt.Println("4. For LINK POSTS: Add to regular queue")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))
}