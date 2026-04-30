// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package appinfo

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-github/v76/github"
)

func TestDiscoveryClient_GetUpdatesFromRepo_LiveServer(t *testing.T) {
	if os.Getenv("SOLO_RUN_LIVE_UPDATE_TESTS") != "1" {
		t.Skip("skipping live update discovery test; set SOLO_RUN_LIVE_UPDATE_TESTS=1 to enable")
	}

	dc := InitDiscoveryCient()

	response, err := dc.GetUpdatesFromRepo("0.1.0-rc1", true)
	if err != nil {
		t.Fatalf("GetUpdatesFromRepo failed: %v", err)
	}
	if response == nil {
		t.Fatal("GetUpdatesFromRepo returned nil response")
	}
	if response.Release == nil {
		t.Fatal("GetUpdatesFromRepo returned no releases")
	}
}

func TestFindLatestReleaseIgnoresPrereleaseForStableCurrent(t *testing.T) {
	releases := []GitHubRelease{
		{TagName: "v1.4.1-rc2", PreRelease: true},
		{TagName: "v1.4.0", PreRelease: false},
		{TagName: "v1.3.9", PreRelease: false},
	}

	got := findLatestRelease(releases, "v1.3.0", false)
	if got == nil {
		t.Fatal("expected an update release")
	}
	if got.TagName != "v1.4.0" {
		t.Fatalf("expected stable update, got %s", got.TagName)
	}
}

func TestFindLatestReleaseIncludesPrereleaseForPrereleaseCurrent(t *testing.T) {
	releases := []GitHubRelease{
		{TagName: "v1.4.0-rc2", PreRelease: true},
		{TagName: "v1.4.0-rc1", PreRelease: true},
	}

	got := findLatestRelease(releases, "v1.4.0-rc1", false)
	if got == nil {
		t.Fatal("expected an update release")
	}
	if got.TagName != "v1.4.0-rc2" {
		t.Fatalf("expected latest prerelease, got %s", got.TagName)
	}
}

func TestFindLatestReleasePicksNewestStableOverOlderPrerelease(t *testing.T) {
	releases := []GitHubRelease{
		{TagName: "v1.5.0-rc1", PreRelease: true},
		{TagName: "v1.4.9", PreRelease: false},
		{TagName: "v1.5.0", PreRelease: false},
	}

	got := findLatestRelease(releases, "v1.4.0", false)
	if got == nil {
		t.Fatal("expected an update release")
	}
	if got.TagName != "v1.5.0" {
		t.Fatalf("expected latest stable release, got %s", got.TagName)
	}
}

func TestFindLatestReleaseIncludesPrereleaseWhenToggleEnabled(t *testing.T) {
	releases := []GitHubRelease{
		{TagName: "v1.5.0-rc1", PreRelease: true},
		{TagName: "v1.4.9", PreRelease: false},
	}

	got := findLatestRelease(releases, "v1.4.0", true)
	if got == nil {
		t.Fatal("expected an update release")
	}
	if got.TagName != "v1.5.0-rc1" {
		t.Fatalf("expected prerelease update when toggle enabled, got %s", got.TagName)
	}
}

func TestToGitHubReleaseIncludesHTMLURL(t *testing.T) {
	release, ok := toGitHubRelease(&github.RepositoryRelease{
		TagName: github.Ptr("v1.2.3"),
		HTMLURL: github.Ptr("https://github.com/raml-dev/solo/releases/tag/v1.2.3"),
	})
	if !ok {
		t.Fatal("expected release conversion to succeed")
	}
	if release.HTMLURL != "https://github.com/raml-dev/solo/releases/tag/v1.2.3" {
		t.Fatalf("unexpected html url: %s", release.HTMLURL)
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
			currentVersion: "v1.4.0-rc1",
			latestVersion:  "v1.4.0-rc3",
			want:           true,
		},
		{
			name:           "current equal latest",
			currentVersion: "v1.4.0-rc3",
			latestVersion:  "v1.4.0-rc3",
			want:           false,
		},
		{
			name:           "current newer than latest",
			currentVersion: "v1.5.0-rc1",
			latestVersion:  "v1.4.9-rc8",
			want:           false,
		},
		{
			name:           "stable beats prerelease on same version",
			currentVersion: "v1.4.0-rc8",
			latestVersion:  "v1.4.0",
			want:           true,
		},
		{
			name:           "empty current means update available",
			currentVersion: "",
			latestVersion:  "v1.4.0-rc3",
			want:           true,
		},
		{
			name:           "dev always updates to latest",
			currentVersion: "dev",
			latestVersion:  "v9.9.9-rc1",
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

func TestShouldIncludePrereleases(t *testing.T) {
	cases := []struct {
		name                     string
		currentVersion           string
		allowConfiguredPreleases bool
		want                     bool
	}{
		{name: "stable current keeps prereleases disabled by default", currentVersion: "v1.4.0", allowConfiguredPreleases: false, want: false},
		{name: "stable current includes prereleases when toggle enabled", currentVersion: "v1.4.0", allowConfiguredPreleases: true, want: true},
		{name: "prerelease current always includes prereleases", currentVersion: "v1.4.0-rc1", allowConfiguredPreleases: false, want: true},
		{name: "dev current always includes prereleases", currentVersion: "dev", allowConfiguredPreleases: false, want: true},
		{name: "empty current always includes prereleases", currentVersion: "", allowConfiguredPreleases: false, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldIncludePrereleases(tc.currentVersion, tc.allowConfiguredPreleases)
			if got != tc.want {
				t.Fatalf("shouldIncludePrereleases(%q, %v) = %v, want %v", tc.currentVersion, tc.allowConfiguredPreleases, got, tc.want)
			}
		})
	}
}

func TestNormalizedReleaseTimePrefersUpdatedAt(t *testing.T) {
	createdAt := time.Date(2026, 4, 1, 8, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Hour)

	got := normalizedReleaseTime(GitHubRelease{
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	})
	if !got.Equal(updatedAt) {
		t.Fatalf("expected updated time, got %s", got)
	}
}
