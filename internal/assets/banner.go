package assets

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// FindBannerFile returns the absolute path and filename of the first banner image
// found in dir. Supported filenames are banner.jpg, banner.jpeg, and banner.png.
// If no banner is found, it returns empty strings and a nil error.
func FindBannerFile(dir string) (string, string, error) {
	candidates := []string{"banner.jpg", "banner.jpeg", "banner.png"}
	for _, name := range candidates {
		fullPath := filepath.Join(dir, name)
		_, err := os.Stat(fullPath)
		if err == nil {
			return fullPath, name, nil
		}
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		return "", "", err
	}
	return "", "", nil
}

