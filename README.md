# Distribute - Content Distribution CLI

A Go CLI tool that automates content distribution from structured Markdown files to multiple platforms:
- Astro blog (file copy)
- ConvertKit newsletter (draft creation)
- Buffer social media scheduler (with platform-specific formatting and threading)

## Installation

```bash
go mod tidy
go build -o distribute
```

## Configuration

Set the following environment variables:

```bash
# Required
export CONVERTKIT_API_SECRET="your-convertkit-api-key"
export ASTRO_CONTENT_PATH="/path/to/your/astro/content/blog"
export BLOG_BASE_URL="https://yourblog.com"
export BUFFER_ACCESS_TOKEN="your-buffer-access-token"
export BUFFER_PROFILE_IDS="profile1,profile2,profile3"

# Optional - Map Buffer profile IDs to platforms
export BUFFER_TWITTER_PROFILE_ID="your-twitter-profile-id"
export BUFFER_LINKEDIN_PROFILE_ID="your-linkedin-profile-id"  
export BUFFER_BLUESKY_PROFILE_ID="your-bluesky-profile-id"
```

## Usage

```bash
# Run in dry-run mode (recommended first)
go run main.go --dry-run --file path/to/your-post.md

# Or with the compiled binary
./distribute --dry-run --file path/to/your-post.md

# Run for real (makes API calls and copies files)
./distribute --file path/to/your-post.md
```

## Markdown File Format

Your content files must follow this exact structure:

```markdown
---
title: "Your Post Title"
publishDate: "2024-01-15T09:00:00-07:00"
newsletterSubject: "📧 Your newsletter subject line"
tags: ["Tag1", "Tag2", "Tag3"]
socialMediaHashtags: "#Hashtag1 #Hashtag2"
contentID: "unique-id-123"
---

Your main thought piece content goes here. 
This becomes the blog post body and newsletter intro.

<!--LINKS_SEPARATOR-->

### First Link Title

- **Title:** The actual title of the linked article
- **URL:** https://example.com/article
- **MyTake:** Your commentary on why this link matters.

---

### Second Link Title

- **Title:** Another great article
- **URL:** https://example.com/another
- **MyTake:** Your thoughts on this resource.
```

## Features

### Platform-Specific Character Limits
- Twitter/X: 280 characters (automatic threading for longer posts)
- LinkedIn: 3000 characters
- Bluesky: 300 characters (automatic threading)

### Smart Scheduling
- Main thought piece post: Added to top of Buffer queue
- Link posts: Added to regular queue order

### Dry Run Mode
Shows exactly what would be done without making any API calls or file changes. Perfect for testing your content format.

## Getting Buffer Profile IDs

1. Get your Buffer access token from https://buffer.com/developers/api
2. Make this API call to get your profile IDs:
```bash
curl "https://api.bufferapp.com/1/profiles.json?access_token=YOUR_TOKEN"
```
3. Note the `id` field for each social account you want to use

## Testing

Two test files are provided:
- `testdata/simple.md` - Basic post that fits within character limits
- `testdata/threading.md` - Longer content that triggers threading on Twitter/Bluesky

## GitHub Actions Integration

Example workflow:

```yaml
name: Distribute Content

on:
  push:
    paths:
      - 'content/posts/*.md'

jobs:
  distribute:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build distribute CLI
        run: |
          cd distribute
          go build -o distribute
      
      - name: Distribute content
        env:
          CONVERTKIT_API_SECRET: ${{ secrets.CONVERTKIT_API_SECRET }}
          ASTRO_CONTENT_PATH: ./src/content/blog
          BLOG_BASE_URL: https://yourblog.com
          BUFFER_ACCESS_TOKEN: ${{ secrets.BUFFER_ACCESS_TOKEN }}
          BUFFER_PROFILE_IDS: ${{ secrets.BUFFER_PROFILE_IDS }}
        run: |
          ./distribute/distribute --file ${{ github.event.head_commit.modified[0] }}
```