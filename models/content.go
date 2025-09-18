package models

type PostMetadata struct {
	Title               string   `yaml:"title"`
	PublishDate         string   `yaml:"publishDate"`
	NewsletterSubject   string   `yaml:"newsletterSubject"`
	Tags                []string `yaml:"tags"`
	SocialMediaHashtags string   `yaml:"socialMediaHashtags"`
	ContentID           string   `yaml:"contentID"`
}

type CuratedLink struct {
	Title   string
	URL     string
	MyTake  string
	Keyword string // e.g., "link", "video", etc.
}

type PrintItem struct {
	Title       string
	Description string
	ImageURL    string
	LinkURL     string
}

type Content struct {
	Metadata      PostMetadata
	ThoughtPiece  string
	Links         []CuratedLink
	PrintItems    []PrintItem // For "What to Print" section
	FooterContent string      // For author bio/about section
	OriginalPath  string
}

const (
	TwitterLimit  = 280
	LinkedInLimit = 3000
	BlueskyLimit  = 300
)

type Platform string

const (
	PlatformTwitter  Platform = "twitter"
	PlatformLinkedIn Platform = "linkedin"
	PlatformBluesky  Platform = "bluesky"
)

func GetCharLimit(platform Platform) int {
	switch platform {
	case PlatformTwitter:
		return TwitterLimit
	case PlatformLinkedIn:
		return LinkedInLimit
	case PlatformBluesky:
		return BlueskyLimit
	default:
		return TwitterLimit
	}
}
