// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package appinfo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v76/github"
)

const (
	defaultRepositoryURL = "https://github.com/raml-dev/solo"
	releasesPerPage      = 20
	checksumsAssetName   = "SHA256SUMS"
)

type DiscoveryClient struct {
	client   *http.Client
	ghClient *github.Client
	owner    string
	repo     string
}

type GitHubResponse struct {
	Release *GitHubRelease
}

type GitHubRelease struct {
	Assets     []Asset   `json:"assets"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	HTMLURL    string    `json:"html_url"`
	UpdatedAt  time.Time `json:"updated_at"`
	Name       string    `json:"name"`
	TagName    string    `json:"tag_name"`
	PreRelease bool      `json:"prerelease"`
}

type Asset struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
	Url   string `json:"browser_download_url"`
}

// SuggestedAssetName returns the platform-compatible asset filename selected by
// the same logic used during download. It returns an empty string when a
// suitable asset cannot be determined.
func SuggestedAssetName(info *GitHubResponse) string {
	if info == nil || info.Release == nil {
		return ""
	}

	asset := selectAsset(info.Release.Assets)
	if asset == nil {
		return ""
	}

	return strings.TrimSpace(asset.Name)
}

func InitDiscoveryCient() *DiscoveryClient {
	return InitDiscoveryClient("")
}

func InitDiscoveryClient(repositoryURL string) *DiscoveryClient {
	owner, repo := parseRepository(repositoryURL)
	httpClient := &http.Client{Timeout: 30 * time.Second}

	return &DiscoveryClient{
		client:   httpClient,
		ghClient: github.NewClient(httpClient),
		owner:    owner,
		repo:     repo,
	}
}

/*
This function determines whether a newer release exists for the current
version and returns the most recent eligible release.
*/
func (dc *DiscoveryClient) GetUpdatesFromRepo(currentVersion string, includePrereleases bool) (*GitHubResponse, error) {
	if dc == nil {
		return nil, errors.New("discovery client not initialized")
	}

	releases, _, err := dc.ghClient.Repositories.ListReleases(
		context.Background(),
		dc.owner,
		dc.repo,
		&github.ListOptions{PerPage: releasesPerPage},
	)
	if err != nil {
		return nil, err
	}

	candidates := make([]GitHubRelease, 0, len(releases))
	for _, release := range releases {
		if release == nil {
			continue
		}

		converted, ok := toGitHubRelease(release)
		if !ok {
			continue
		}

		candidates = append(candidates, converted)
	}

	latest := findLatestRelease(candidates, currentVersion, includePrereleases)
	if latest == nil {
		return nil, nil
	}

	return &GitHubResponse{Release: latest}, nil
}

func (dc *DiscoveryClient) DownloadAssets(info *GitHubResponse, currentVersion string) (string, error) {
	return dc.downloadAssets(info, currentVersion, "")
}

func (dc *DiscoveryClient) DownloadAssetsToPath(info *GitHubResponse, currentVersion, destinationPath string) (string, error) {
	if strings.TrimSpace(destinationPath) == "" {
		return "", errors.New("destination path is required")
	}

	return dc.downloadAssets(info, currentVersion, destinationPath)
}

func (dc *DiscoveryClient) downloadAssets(info *GitHubResponse, currentVersion, destinationPath string) (string, error) {
	if info == nil || info.Release == nil {
		return "", errors.New("no releases available")
	}

	if !isCurrentVersionOlder(currentVersion, info.Release.TagName) {
		return "", nil
	}

	asset := selectAsset(info.Release.Assets)
	if asset == nil {
		return "", fmt.Errorf("no compatible asset found for runtime %s/%s (%d-bit)", runtime.GOOS, runtime.GOARCH, strconv.IntSize)
	}

	expectedChecksums, err := dc.fetchReleaseChecksums(info.Release.Assets)
	if err != nil {
		return "", err
	}

	expectedChecksum := expectedChecksums[asset.Name]
	if expectedChecksum == "" {
		return "", fmt.Errorf("missing checksum for asset %s", asset.Name)
	}

	outputPath := destinationPath
	if outputPath == "" {
		outputPath = asset.Name
	}

	if err := dc.downloadAssetToPath(asset.Url, outputPath, expectedChecksum); err != nil {
		return "", err
	}

	return info.Release.Body, nil
}

func (dc *DiscoveryClient) fetchReleaseChecksums(assets []Asset) (map[string]string, error) {
	checksumAsset := findChecksumsAsset(assets)
	if checksumAsset == nil {
		return nil, errors.New("release checksums asset not found")
	}

	body, err := dc.downloadText(checksumAsset.Url)
	if err != nil {
		return nil, err
	}

	checksums, err := parseChecksums(body)
	if err != nil {
		return nil, err
	}
	if len(checksums) == 0 {
		return nil, errors.New("release checksums are empty")
	}

	return checksums, nil
}

func (dc *DiscoveryClient) downloadText(url string) (string, error) {
	resp, err := dc.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := expectHTTPSuccess(resp); err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (dc *DiscoveryClient) downloadAssetToPath(downloadURL, destinationPath, expectedChecksum string) (err error) {
	resp, err := dc.client.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := expectHTTPSuccess(resp); err != nil {
		return err
	}

	destinationPath = filepath.Clean(destinationPath)
	parentDir := filepath.Dir(destinationPath)
	if mkErr := os.MkdirAll(parentDir, 0o755); mkErr != nil {
		return mkErr
	}

	tempFile, err := os.CreateTemp(parentDir, filepath.Base(destinationPath)+".*.part")
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	defer func() {
		closeErr := tempFile.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(tempPath)
		}
	}()

	hasher := sha256.New()
	if _, err = io.Copy(io.MultiWriter(tempFile, hasher), resp.Body); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", filepath.Base(destinationPath), expectedChecksum, actualChecksum)
	}

	if chmodErr := maybeMakeExecutable(tempFile.Name()); chmodErr != nil {
		return chmodErr
	}

	if err := tempFile.Close(); err != nil {
		tempFile = nil
		return err
	}
	tempFile = nil

	if removeErr := os.Remove(destinationPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return removeErr
	}

	return os.Rename(tempPath, destinationPath)
}

func toGitHubRelease(release *github.RepositoryRelease) (GitHubRelease, bool) {
	tagName := strings.TrimSpace(release.GetTagName())
	if tagName == "" {
		return GitHubRelease{}, false
	}

	assets := make([]Asset, 0, len(release.Assets))
	for _, asset := range release.Assets {
		assets = append(assets, Asset{
			ID:    asset.GetID(),
			Name:  asset.GetName(),
			State: asset.GetState(),
			Url:   asset.GetBrowserDownloadURL(),
		})
	}

	return GitHubRelease{
		Assets:     assets,
		Body:       release.GetBody(),
		CreatedAt:  release.GetCreatedAt().Time,
		HTMLURL:    release.GetHTMLURL(),
		UpdatedAt:  release.GetPublishedAt().Time,
		Name:       release.GetName(),
		TagName:    tagName,
		PreRelease: release.GetPrerelease(),
	}, true
}

func findLatestRelease(releases []GitHubRelease, currentVersion string, includePrereleases bool) *GitHubRelease {
	includePrereleases = shouldIncludePrereleases(currentVersion, includePrereleases)

	var latest *GitHubRelease
	for i := range releases {
		release := releases[i]
		if strings.TrimSpace(release.TagName) == "" {
			continue
		}
		if release.PreRelease && !includePrereleases {
			continue
		}
		if !isCurrentVersionOlder(currentVersion, release.TagName) {
			continue
		}
		if latest == nil || compareVersionStrings(release.TagName, latest.TagName) > 0 {
			chosen := release
			latest = &chosen
		}
	}

	return latest
}

func findChecksumsAsset(assets []Asset) *Asset {
	for i := range assets {
		if strings.EqualFold(strings.TrimSpace(assets[i].Name), checksumsAssetName) {
			return &assets[i]
		}
	}

	return nil
}

func expectHTTPSuccess(resp *http.Response) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("request failed: %s - %s", resp.Status, strings.TrimSpace(string(body)))
}

func maybeMakeExecutable(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}

	if runtime.GOOS == "darwin" {
		return nil
	}

	return os.Chmod(path, 0o755)
}
