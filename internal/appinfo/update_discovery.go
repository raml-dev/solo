// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package appinfo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	HTMLURL    string    `json:"html_url"`
	UpdatedAt  time.Time `json:"updated_at"`
	Name       string    `json:"name"`
	TagName    string    `json:"tag_name"`
	PreRelease bool      `json:"prerelease"`
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

func toGitHubRelease(release *github.RepositoryRelease) (GitHubRelease, bool) {
	tagName := strings.TrimSpace(release.GetTagName())
	if tagName == "" {
		return GitHubRelease{}, false
	}

	return GitHubRelease{
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

func expectHTTPSuccess(resp *http.Response) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("request failed: %s - %s", resp.Status, strings.TrimSpace(string(body)))
}
