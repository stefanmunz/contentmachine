package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHomeDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	_ = os.Setenv("HOME", home)

	got := expandHomeDir("~/some/path")
	want := filepath.Join(home, "some", "path")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	got = expandHomeDir("~")
	if got != home {
		t.Fatalf("expected %q, got %q", home, got)
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"stefanmunz.com", "https://stefanmunz.com"},
		{"stefanmunz.com/", "https://stefanmunz.com"},
		{"https://stefanmunz.com", "https://stefanmunz.com"},
		{"https://stefanmunz.com/", "https://stefanmunz.com"},
		{"http://stefanmunz.com/", "http://stefanmunz.com"},
		{"", ""},
		{"   stefanmunz.com  ", "https://stefanmunz.com"},
	}

	for _, tc := range cases {
		got := normalizeBaseURL(tc.in)
		if got != tc.want {
			t.Fatalf("normalizeBaseURL(%q): expected %q, got %q", tc.in, tc.want, got)
		}
	}
}
