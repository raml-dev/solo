// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package appinfo

import (
	"os"
	"path/filepath"
	"runtime"
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

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, SuggestedAssetName(response))
	if outputPath == tempDir {
		t.Fatal("SuggestedAssetName returned empty output filename")
	}

	body, err := dc.DownloadAssetsToPath(response, "", outputPath)
	if err != nil {
		t.Fatalf("DownloadAssetsToPath failed: %v", err)
	}
	if body == "" {
		t.Fatal("DownloadAssetsToPath returned empty release body")
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("downloaded file stat failed: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("downloaded file is empty")
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

func TestSelectAssetUsesWorkflowNames(t *testing.T) {
	assets := []Asset{
		{Name: "solo-windows-amd64.exe"},
		{Name: "solo-windows-arm64.exe"},
		{Name: "solo-linux-amd64"},
		{Name: "solo-linux-arm64"},
		{Name: "solo-darwin-amd64.dmg"},
		{Name: "solo-darwin-arm64.dmg"},
	}

	expected := expectedAssetName(runtime.GOOS, runtime.GOARCH)
	if expected == "" {
		t.Skipf("runtime %s/%s is not supported by release workflow", runtime.GOOS, runtime.GOARCH)
	}

	got := selectAsset(assets)
	if got == nil {
		t.Fatal("expected selected asset")
	}
	if got.Name != expected {
		t.Fatalf("expected %s, got %s", expected, got.Name)
	}
}

func TestExpectedAssetNamesFollowReleaseContract(t *testing.T) {
	cases := []struct {
		goos   string
		goarch string
		want   string
	}{
		{goos: "windows", goarch: "amd64", want: "solo-windows-amd64.exe"},
		{goos: "windows", goarch: "arm64", want: "solo-windows-arm64.exe"},
		{goos: "linux", goarch: "amd64", want: "solo-linux-amd64"},
		{goos: "linux", goarch: "arm64", want: "solo-linux-arm64"},
		{goos: "darwin", goarch: "amd64", want: "solo-darwin-amd64.dmg"},
		{goos: "darwin", goarch: "arm64", want: "solo-darwin-arm64.dmg"},
	}

	for _, tc := range cases {
		t.Run(tc.goos+"-"+tc.goarch, func(t *testing.T) {
			got := expectedAssetName(tc.goos, tc.goarch)
			if got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestSelectAssetIgnoresStableOnlyPackageArtifacts(t *testing.T) {
	assets := []Asset{
		{Name: "solo_1.2.3_amd64.deb"},
		{Name: "solo_1.2.3_arm64.deb"},
		{Name: "solo-1.2.3-1.x86_64.rpm"},
		{Name: "solo-1.2.3-1.aarch64.rpm"},
		{Name: "solo-1.2.3-1-x86_64.pkg.tar.zst"},
		{Name: "solo-1.2.3-1-aarch64.pkg.tar.zst"},
		{Name: "solo-windows-amd64.exe"},
		{Name: "solo-windows-arm64.exe"},
		{Name: "solo-linux-amd64"},
		{Name: "solo-linux-arm64"},
		{Name: "solo-darwin-amd64.dmg"},
		{Name: "solo-darwin-arm64.dmg"},
	}

	expected := expectedAssetName(runtime.GOOS, runtime.GOARCH)
	if expected == "" {
		t.Skipf("runtime %s/%s is not supported by release workflow", runtime.GOOS, runtime.GOARCH)
	}

	got := selectAsset(assets)
	if got == nil {
		t.Fatal("expected selected raw asset")
	}
	if got.Name != expected {
		t.Fatalf("expected raw asset %s, got %s", expected, got.Name)
	}
}

func TestParseChecksums(t *testing.T) {
	checksums, err := parseChecksums(`
abc123  solo-linux-amd64
def456 *solo-darwin-arm64.dmg
`)
	if err != nil {
		t.Fatalf("parseChecksums failed: %v", err)
	}

	if checksums["solo-linux-amd64"] != "abc123" {
		t.Fatalf("unexpected checksum for linux asset: %#v", checksums)
	}
	if checksums["solo-darwin-arm64.dmg"] != "def456" {
		t.Fatalf("unexpected checksum for darwin asset: %#v", checksums)
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

func TestFetchReleaseChecksumsRequiresChecksumsAsset(t *testing.T) {
	dc := InitDiscoveryClient("")

	_, err := dc.fetchReleaseChecksums([]Asset{{Name: "solo-linux-amd64"}})
	if err == nil {
		t.Fatal("expected missing checksums asset to fail")
	}
	if err.Error() != "release checksums asset not found" {
		t.Fatalf("unexpected error: %v", err)
	}
}
