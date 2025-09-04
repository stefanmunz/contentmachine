package config

import (
	"distribute/models"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	KitAPIKey           string // Kit v4 API key
	AstroContentPath    string
	BlogBaseURL         string
	BufferAccessToken   string
	BufferProfileIDs    []string
	ProfilePlatformMap  map[string]models.Platform
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Kit v4 API key is required
	cfg.KitAPIKey = os.Getenv("KIT_API_KEY")
	if cfg.KitAPIKey == "" {
		return nil, fmt.Errorf("KIT_API_KEY environment variable is required")
	}

	cfg.AstroContentPath = os.Getenv("ASTRO_CONTENT_PATH")
	if cfg.AstroContentPath == "" {
		return nil, fmt.Errorf("ASTRO_CONTENT_PATH environment variable is required")
	}

	cfg.BlogBaseURL = os.Getenv("BLOG_BASE_URL")
	if cfg.BlogBaseURL == "" {
		return nil, fmt.Errorf("BLOG_BASE_URL environment variable is required")
	}

	cfg.BufferAccessToken = os.Getenv("BUFFER_ACCESS_TOKEN")
	if cfg.BufferAccessToken == "" {
		return nil, fmt.Errorf("BUFFER_ACCESS_TOKEN environment variable is required")
	}

	profileIDsStr := os.Getenv("BUFFER_PROFILE_IDS")
	if profileIDsStr == "" {
		return nil, fmt.Errorf("BUFFER_PROFILE_IDS environment variable is required")
	}
	cfg.BufferProfileIDs = strings.Split(profileIDsStr, ",")

	// Hardcoded platform mapping - you'll need to update these with your actual profile IDs
	cfg.ProfilePlatformMap = map[string]models.Platform{
		"TWITTER_PROFILE_ID":  models.PlatformTwitter,
		"LINKEDIN_PROFILE_ID": models.PlatformLinkedIn,
		"BLUESKY_PROFILE_ID":  models.PlatformBluesky,
	}

	// Optional: Allow overriding the platform map via environment variables
	if twitterID := os.Getenv("BUFFER_TWITTER_PROFILE_ID"); twitterID != "" {
		cfg.ProfilePlatformMap[twitterID] = models.PlatformTwitter
	}
	if linkedinID := os.Getenv("BUFFER_LINKEDIN_PROFILE_ID"); linkedinID != "" {
		cfg.ProfilePlatformMap[linkedinID] = models.PlatformLinkedIn
	}
	if blueskyID := os.Getenv("BUFFER_BLUESKY_PROFILE_ID"); blueskyID != "" {
		cfg.ProfilePlatformMap[blueskyID] = models.PlatformBluesky
	}

	return cfg, nil
}