// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only
package appinfo

/*
  NOTE:
    this module is a tempi implementation just for the beta version of this app (Solo).
    It will be totally (more or less) replaced by new implementation when the beta version
    will be over.

*/

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v76/github"
	"github.com/google/uuid"
	"github.com/matstech/aegis-go/client"
)

// This var is used for storing custom configuration endpoints and secret
// for aegis server interaction
//
//go:embed update_config.json
var embeddedUpdateConfig []byte

var (
	discoveryConfigOnce sync.Once
	discoveryConfigData discoveryConfig
	discoveryConfigErr  error
)

type DiscoveryClient struct {
	client   *http.Client
	endpoint string
	ghClient *github.Client
}

type GitHubResponse struct {
	Release *GitHubRelease
}

type GitHubRelease struct {
	Assets     []Asset   `json:"assets"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
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

type discoveryConfig struct {
	Endpoint string `json:"endpoint"`
	Kid      string `json:"kid"`
	Secret   string `json:"secret"`
}

type parsedVersion struct {
	major      int
	minor      int
	patch      int
	prerelease string
}

// Aegis-go client round tripper
type roundTripperFunc func(*http.Request) (*http.Response, error)

// Basic round-trip function
func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// SuggestedAssetName returns the platform-compatible asset filename selected by
// the same logic used during download. It returns an empty string when a
// suitable asset cannot be determined.
func SuggestedAssetName(info *GitHubResponse) string {
	if info == nil {
		return ""
	}

	asset := selectAsset(info.Release.Assets)
	if asset == nil {
		return ""
	}

	return strings.TrimSpace(asset.Name)
}

func InitDiscoveryCient() *DiscoveryClient {
	cfg, err := loadDiscoveryConfig()
	if err != nil {
		slog.Error("Failed to load embedded update discovery configuration", "error", err)
	}

	httpClient := &http.Client{
		Transport: client.NewTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return http.DefaultTransport.RoundTrip(req)
		}), client.Config{
			Kid:    cfg.Kid,
			Secret: cfg.Secret,
		}),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return newDiscoveryClient(cfg.Endpoint, httpClient)
}

/*
	  This function determines the effective update availability of a new release over the currentVersion.
		-- Definition of "update availability" --
		   `currentVersion` must appear in the list of versions (versions must be immutable) and its timestamp (the `created_at` field will suffice)
		   must be earlier than that of the others. The latest version will be returned to the frontend
*/
func (dc *DiscoveryClient) GetUpdatesFromRepo(currentVersion string) (*GitHubResponse, error) {
	owner, repo, perPage, err := parseRepoEndpoint(dc.endpoint)
	if err != nil {
		return nil, err
	}

	options := &github.ListOptions{PerPage: perPage}
	releases, _, err := dc.ghClient.Repositories.ListReleases(context.Background(), owner, repo, options)
	if err != nil {
		return nil, err
	}

	// Check if the currentVersion appears in the release list
	var cv *github.RepositoryRelease
	for _, release := range releases {
		if *release.Name == currentVersion {
			cv = release
			break
		}
	}

	if cv == nil {
		return nil, fmt.Errorf("current version %s of Solo does not appear in the releases and the update cannot be determined", currentVersion)
	}

	result := make([]GitHubRelease, 0, len(releases))
	for _, release := range releases {
		if release == nil {
			continue
		}

		if *release.Prerelease && release.CreatedAt.Time.After(cv.CreatedAt.Time) {

			assets := make([]Asset, 0, len(release.Assets))
			for _, asset := range release.Assets {
				assets = append(assets, Asset{
					ID:    asset.GetID(),
					Name:  asset.GetName(),
					State: asset.GetState(),
					Url:   asset.GetBrowserDownloadURL(),
				})
			}

			result = append(result, GitHubRelease{
				Assets:     assets,
				Body:       release.GetBody(),
				CreatedAt:  release.GetCreatedAt().Time,
				UpdatedAt:  release.GetPublishedAt().Time,
				Name:       release.GetName(),
				TagName:    release.GetTagName(),
				PreRelease: release.GetPrerelease(),
			})
		}
	}

	if len(result) <= 0 {
		return nil, nil
	}

	// maybe useless but not so expensive in this case, just a "pac"
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return &GitHubResponse{Release: &result[0]}, nil
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

//------ Utilities and tools

func (dc *DiscoveryClient) newGetUpdatesRequest() (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, dc.endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-CorrelationId", uuid.NewString())

	return req, nil
}

func (dc *DiscoveryClient) newGetAssetsRequest(endpoint string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-CorrelationId", uuid.NewString())
	req.Header.Set("Accept", "application/octet-stream")

	return req, nil
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

	owner, repo, _, err := parseRepoEndpoint(dc.endpoint)
	if err != nil {
		return "", err
	}

	assetURL, err := buildProxyAssetURL(dc.endpoint, owner, repo, asset.ID)
	if err != nil {
		return "", err
	}
	req, err := dc.newGetAssetsRequest(assetURL)
	if err != nil {
		return "", err
	}
	resp, err := dc.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	outputPath := asset.Name
	if destinationPath != "" {
		outputPath = destinationPath
	}

	fileout, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer fileout.Close()

	if resp.StatusCode >= http.StatusMultipleChoices && resp.StatusCode < http.StatusBadRequest {
		location := resp.Header.Get("Location")
		if location == "" {
			return "", fmt.Errorf("asset download redirected by proxy without location: %s", resp.Status)
		}

		redirectReq, redirectErr := dc.newGetAssetsRequest(location)
		if redirectErr != nil {
			return "", redirectErr
		}
		redirectResp, redirectDoErr := dc.client.Do(redirectReq)
		if redirectDoErr != nil {
			return "", redirectDoErr
		}
		defer redirectResp.Body.Close()
		if redirectResp.StatusCode < http.StatusOK || redirectResp.StatusCode >= http.StatusMultipleChoices {
			body, _ := io.ReadAll(redirectResp.Body)
			return "", fmt.Errorf("asset redirected download failed: %s - %s", redirectResp.Status, string(body))
		}

		_, copyErr := io.Copy(fileout, redirectResp.Body)
		if copyErr != nil {
			return "", copyErr
		}
		return info.Release.Body, nil
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("asset download failed: %s - %s", resp.Status, string(body))
	}
	_, copyErr := io.Copy(fileout, resp.Body)
	if copyErr != nil {
		return "", copyErr
	}

	return info.Release.Body, nil
}

func (dc *DiscoveryClient) doRequest(req *http.Request) string {
	resp, err := dc.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body)
}
