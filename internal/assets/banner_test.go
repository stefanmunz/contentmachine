package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindBannerFile_PrefersJpgOverJpegOverPng(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "banner.png"), []byte("x"), 0644); err != nil {
		t.Fatalf("write banner.png: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "banner.jpeg"), []byte("x"), 0644); err != nil {
		t.Fatalf("write banner.jpeg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "banner.jpg"), []byte("x"), 0644); err != nil {
		t.Fatalf("write banner.jpg: %v", err)
	}

	path, name, err := FindBannerFile(dir)
	if err != nil {
		t.Fatalf("FindBannerFile returned error: %v", err)
	}
	if name != "banner.jpg" {
		t.Fatalf("expected banner.jpg, got %q", name)
	}
	if want := filepath.Join(dir, "banner.jpg"); path != want {
		t.Fatalf("expected path %q, got %q", want, path)
	}
}

func TestFindBannerFile_ReturnsEmptyWhenMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path, name, err := FindBannerFile(dir)
	if err != nil {
		t.Fatalf("FindBannerFile returned error: %v", err)
	}
	if path != "" || name != "" {
		t.Fatalf("expected empty result, got path=%q name=%q", path, name)
	}
}

