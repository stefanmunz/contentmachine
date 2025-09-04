package main

import (
	"distribute/config"
	"distribute/handlers"
	"distribute/models"
	"distribute/parser"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	filePath string
	dryRun   bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "distribute",
		Short: "Distribute content to blog, newsletter, and social media",
		Long: `A CLI tool that parses structured Markdown files and distributes
content to Astro blog, ConvertKit newsletter, and Buffer social media scheduler.`,
		Run: run,
	}

	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the markdown file (required)")
	rootCmd.MarkFlagRequired("file")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show what would be done without making changes")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables only")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		if dryRun {
			// In dry-run mode, create a mock config for demonstration
			log.Println("WARNING: Running in dry-run mode with mock configuration")
			cfg = &config.Config{
				KitAPIKey:          "mock-kit-v4-api-key",
				AstroContentPath:   "/path/to/astro/content/blog",
				BlogBaseURL:        "https://myblog.com",
				BufferAccessToken:  "mock-buffer-token",
				BufferProfileIDs:   []string{"TWITTER_PROFILE_ID", "LINKEDIN_PROFILE_ID", "BLUESKY_PROFILE_ID"},
				ProfilePlatformMap: map[string]models.Platform{
					"TWITTER_PROFILE_ID":  models.PlatformTwitter,
					"LINKEDIN_PROFILE_ID": models.PlatformLinkedIn,
					"BLUESKY_PROFILE_ID":  models.PlatformBluesky,
				},
			}
		} else {
			log.Fatalf("Error loading configuration: %v", err)
		}
	}

	// Parse the markdown file
	content, err := parser.ParseMarkdownFile(filePath)
	if err != nil {
		log.Fatalf("Error parsing markdown file: %v", err)
	}

	log.Printf("INFO: Successfully parsed file: %s", filePath)
	log.Printf("INFO: Title: %s", content.Metadata.Title)
	log.Printf("INFO: Found %d curated links", len(content.Links))

	// 1. Handle Astro blog post
	if err := handlers.HandleAstroPost(cfg, content, dryRun); err != nil {
		log.Fatalf("Error handling Astro post: %v", err)
	}

	// 2. Create ConvertKit draft
	if err := handlers.CreateConvertKitDraft(cfg, content, dryRun); err != nil {
		log.Fatalf("Error creating ConvertKit draft: %v", err)
	}

	// 3. Schedule Buffer posts
	if err := handlers.HandleBufferScheduling(cfg, content, dryRun); err != nil {
		log.Fatalf("Error scheduling Buffer posts: %v", err)
	}

	if dryRun {
		log.Println("INFO: Dry run completed successfully")
	} else {
		log.Println("INFO: Content distribution completed successfully")
	}
}