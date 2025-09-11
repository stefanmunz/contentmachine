package config

import (
	"distribute/models"
	"fmt"
	"os"
	"strings"
)

type BlogConfig struct {
	ContentPath string // Path to blog content directory
	RepoPath    string // Path to blog repository
	BaseURL     string // Base URL of the blog
}

type Config struct {
	KitAPIKey           string // Kit v4 API key
	PersonalBlog        BlogConfig // Personal blog (stefanmunz.com)
	OnTreeBlog          BlogConfig // OnTree blog
	BufferAccessToken   string
	BufferProfileIDs    []string
	ProfilePlatformMap  map[string]models.Platform
	// Legacy fields for backward compatibility
	AstroContentPath    string
	BlogBaseURL         string
	BlogRepoPath        string
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Kit v4 API key is optional - will skip Kit integration if not provided
	cfg.KitAPIKey = os.Getenv("KIT_API_KEY")

	// Personal blog configuration (stefanmunz.com)
	cfg.PersonalBlog.ContentPath = os.Getenv("PERSONAL_BLOG_CONTENT_PATH")
	if cfg.PersonalBlog.ContentPath == "" {
		// Fall back to legacy env var
		cfg.PersonalBlog.ContentPath = os.Getenv("ASTRO_CONTENT_PATH")
	}
	if cfg.PersonalBlog.ContentPath == "" {
		return nil, fmt.Errorf("PERSONAL_BLOG_CONTENT_PATH or ASTRO_CONTENT_PATH environment variable is required")
	}

	cfg.PersonalBlog.BaseURL = os.Getenv("PERSONAL_BLOG_BASE_URL")
	if cfg.PersonalBlog.BaseURL == "" {
		// Fall back to legacy env var
		cfg.PersonalBlog.BaseURL = os.Getenv("BLOG_BASE_URL")
	}
	if cfg.PersonalBlog.BaseURL == "" {
		return nil, fmt.Errorf("PERSONAL_BLOG_BASE_URL or BLOG_BASE_URL environment variable is required")
	}

	cfg.PersonalBlog.RepoPath = os.Getenv("PERSONAL_BLOG_REPO_PATH")
	if cfg.PersonalBlog.RepoPath == "" {
		// Fall back to legacy env var
		cfg.PersonalBlog.RepoPath = os.Getenv("BLOG_REPO_PATH")
	}

	// OnTree blog configuration
	cfg.OnTreeBlog.ContentPath = os.Getenv("ONTREE_BLOG_CONTENT_PATH")
	cfg.OnTreeBlog.RepoPath = os.Getenv("ONTREE_BLOG_REPO_PATH")
	cfg.OnTreeBlog.BaseURL = os.Getenv("ONTREE_BLOG_BASE_URL")

	// Set legacy fields for backward compatibility
	cfg.AstroContentPath = cfg.PersonalBlog.ContentPath
	cfg.BlogBaseURL = cfg.PersonalBlog.BaseURL
	cfg.BlogRepoPath = cfg.PersonalBlog.RepoPath

	// Buffer access token is optional - will just display posts if not provided
	cfg.BufferAccessToken = os.Getenv("BUFFER_ACCESS_TOKEN")

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