package appinfo

import (
	"testing"
	"time"
)

func TestDiscoveryClient_GetUpdatesFromRepo_LiveServer(t *testing.T) {
	dc := InitDiscoveryCient()

	response, err := dc.GetUpdatesFromRepo()
	if err != nil {
		t.Fatalf("GetUpdatesFromRepo failed: %v", err)
	}
	if response == nil {
		t.Fatal("GetUpdatesFromRepo returned nil response")
	}
	if len(response.Releases) == 0 {
		t.Fatal("GetUpdatesFromRepo returned no releases")
	}

	s, err := dc.DownloadAssets(response, "")
	if err != nil {
		t.Fatalf("DownloadAssets failed: %v", err)
	}
	if s == "" {
		t.Fatal("DownloadAssets returned empty release body")
	}
}

func TestSelectLatestPrerelease(t *testing.T) {
	releases := []GitHubRelease{
		{TagName: "v1.3.0", PreRelease: false, Assets: []Asset{{ID: 1}}},
		{TagName: "v1.4.0-rc.1", PreRelease: true, Assets: []Asset{{ID: 2}}},
		{TagName: "v1.4.0-rc.3", PreRelease: true, Assets: []Asset{{ID: 3}}},
		{TagName: "v1.4.0-rc.2", PreRelease: true, Assets: []Asset{{ID: 4}}},
	}

	got := selectLatestPrerelease(releases)
	if got == nil {
		t.Fatal("expected a prerelease, got nil")
	}
	if got.TagName != "v1.4.0-rc.3" {
		t.Fatalf("expected latest prerelease v1.4.0-rc.3, got %s", got.TagName)
	}
}

func TestIsCurrentVersionOlder(t *testing.T) {
	cases := []struct {
		name           string
		currentVersion string
		latestVersion  string
		want           bool
	}{
		{
			name:           "current older prerelease",
			currentVersion: "v1.4.0-rc.1",
			latestVersion:  "v1.4.0-rc.3",
			want:           true,
		},
		{
			name:           "current equal latest",
			currentVersion: "v1.4.0-rc.3",
			latestVersion:  "v1.4.0-rc.3",
			want:           false,
		},
		{
			name:           "current newer than latest",
			currentVersion: "v1.5.0-rc.1",
			latestVersion:  "v1.4.9-rc.8",
			want:           false,
		},
		{
			name:           "empty current means update available",
			currentVersion: "",
			latestVersion:  "v1.4.0-rc.3",
			want:           true,
		},
		{
			name:           "dev always updates to latest",
			currentVersion: "dev",
			latestVersion:  "v9.9.9-rc.1",
			want:           true,
		},
		{
			name:           "fallback to lexical compare",
			currentVersion: "build-001",
			latestVersion:  "build-002",
			want:           true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isCurrentVersionOlder(tc.currentVersion, tc.latestVersion)
			if got != tc.want {
				t.Fatalf("isCurrentVersionOlder(%q, %q) = %v, want %v", tc.currentVersion, tc.latestVersion, got, tc.want)
			}
		})
	}
}

func TestSelectLatestPrereleaseFallbackByTime(t *testing.T) {
	now := time.Now()
	releases := []GitHubRelease{
		{
			TagName:    "invalid-tag-a",
			PreRelease: true,
			Assets:     []Asset{{ID: 1}},
			UpdatedAt:  now.Add(-2 * time.Hour),
		},
		{
			TagName:    "invalid-tag-b",
			PreRelease: true,
			Assets:     []Asset{{ID: 2}},
			UpdatedAt:  now.Add(-1 * time.Hour),
		},
	}

	got := selectLatestPrerelease(releases)
	if got == nil {
		t.Fatal("expected latest prerelease, got nil")
	}
	if got.TagName != "invalid-tag-b" {
		t.Fatalf("expected latest by UpdatedAt invalid-tag-b, got %s", got.TagName)
	}
}
