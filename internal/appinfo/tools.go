// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package appinfo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v76/github"
)

func loadDiscoveryConfig() (discoveryConfig, error) {
	discoveryConfigOnce.Do(func() {
		if len(embeddedUpdateConfig) == 0 {
			discoveryConfigErr = errors.New("embedded update configuration is empty")
			return
		}

		var cfg discoveryConfig
		if err := json.Unmarshal(embeddedUpdateConfig, &cfg); err != nil {
			discoveryConfigErr = fmt.Errorf("invalid embedded update configuration: %w", err)
			return
		}
		if strings.TrimSpace(cfg.Endpoint) == "" {
			discoveryConfigErr = errors.New("update endpoint is required")
			return
		}
		if strings.TrimSpace(cfg.Kid) == "" {
			discoveryConfigErr = errors.New("update kid is required")
			return
		}
		if strings.TrimSpace(cfg.Secret) == "" {
			discoveryConfigErr = errors.New("update secret is required")
			return
		}

		discoveryConfigData = cfg
	})

	return discoveryConfigData, discoveryConfigErr
}

func newDiscoveryClient(endpoint string, httpClient *http.Client) *DiscoveryClient {
	ghClient := github.NewClient(httpClient)
	parsedEndpoint, err := url.Parse(endpoint)
	if err == nil {
		ghClient.BaseURL = &url.URL{
			Scheme: parsedEndpoint.Scheme,
			Host:   parsedEndpoint.Host,
			Path:   "/",
		}
	}

	return &DiscoveryClient{
		client:   httpClient,
		endpoint: endpoint,
		ghClient: ghClient,
	}
}

func parseRepoEndpoint(endpoint string) (owner string, repo string, perPage int, err error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", "", 0, err
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[0] != "repos" {
		return "", "", 0, fmt.Errorf("invalid release endpoint path: %s", parsed.Path)
	}

	perPage = 10
	if value := parsed.Query().Get("per_page"); value != "" {
		parsedPerPage, convErr := strconv.Atoi(value)
		if convErr != nil || parsedPerPage <= 0 {
			return "", "", 0, fmt.Errorf("invalid per_page value: %s", value)
		}
		perPage = parsedPerPage
	}

	return parts[1], parts[2], perPage, nil
}

func buildProxyAssetURL(endpoint string, owner string, repo string, assetID int64) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	parsed.Path = fmt.Sprintf("/repos/%s/%s/releases/assets/%d", owner, repo, assetID)
	parsed.RawQuery = ""
	return parsed.String(), nil
}

func normalizedReleaseTime(release GitHubRelease) time.Time {
	if !release.UpdatedAt.IsZero() {
		return release.UpdatedAt
	}
	return release.CreatedAt
}

func isCurrentVersionOlder(currentVersion, latestVersion string) bool {
	current := strings.TrimSpace(currentVersion)
	latest := strings.TrimSpace(latestVersion)

	if strings.EqualFold(current, "dev") {
		return latest != ""
	}

	// If the caller doesn't provide a current version, consider any latest release as newer.
	if current == "" {
		return latest != ""
	}

	return compareVersionStrings(current, latest) < 0
}

func compareVersionStrings(left, right string) int {
	leftVersion, okLeft := parseVersion(left)
	rightVersion, okRight := parseVersion(right)

	switch {
	case okLeft && okRight:
		return compareParsedVersion(leftVersion, rightVersion)
	case okLeft:
		return 1
	case okRight:
		return -1
	default:
		return strings.Compare(strings.TrimSpace(left), strings.TrimSpace(right))
	}
}

func parseVersion(raw string) (parsedVersion, bool) {
	value := strings.TrimSpace(strings.TrimPrefix(raw, "v"))
	if value == "" {
		return parsedVersion{}, false
	}

	base := value
	prerelease := ""
	if idx := strings.Index(base, "+"); idx >= 0 {
		base = base[:idx]
	}
	if idx := strings.Index(base, "-"); idx >= 0 {
		prerelease = base[idx+1:]
		base = base[:idx]
	}

	parts := strings.Split(base, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return parsedVersion{}, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return parsedVersion{}, false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return parsedVersion{}, false
	}

	patch := 0
	if len(parts) == 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return parsedVersion{}, false
		}
	}

	return parsedVersion{
		major:      major,
		minor:      minor,
		patch:      patch,
		prerelease: prerelease,
	}, true
}

func compareParsedVersion(left, right parsedVersion) int {
	if left.major != right.major {
		if left.major > right.major {
			return 1
		}
		return -1
	}
	if left.minor != right.minor {
		if left.minor > right.minor {
			return 1
		}
		return -1
	}
	if left.patch != right.patch {
		if left.patch > right.patch {
			return 1
		}
		return -1
	}

	// Stable release outranks prerelease when core versions match.
	if left.prerelease == "" && right.prerelease != "" {
		return 1
	}
	if left.prerelease != "" && right.prerelease == "" {
		return -1
	}
	return strings.Compare(left.prerelease, right.prerelease)
}

func selectAsset(assets []Asset) *Asset {
	osName := strings.ToLower(runtime.GOOS)
	arch := strings.ToLower(runtime.GOARCH)

	bestScore := -1
	var best *Asset
	for _, asset := range assets {
		if asset.ID == 0 {
			continue
		}

		name := strings.ToLower(asset.Name)
		if name == "" {
			name = strings.ToLower(asset.Url)
		}

		score := 0
		if strings.Contains(name, osName) {
			score += 4
		}
		if strings.Contains(name, arch) {
			score += 5
		}
		if is32BitRuntime() && strings.Contains(name, "32") {
			score += 2
		}
		if !is32BitRuntime() && strings.Contains(name, "64") {
			score += 2
		}
		if runtime.GOOS == "darwin" && strings.Contains(name, ".dmg") {
			score += 2
		}
		if runtime.GOOS == "windows" && strings.Contains(name, ".exe") {
			score += 2
		}
		if runtime.GOOS == "linux" && strings.Contains(name, ".appimage") {
			score += 2
		}

		if score > bestScore {
			bestScore = score
			chosen := asset
			best = &chosen
		}
	}

	return best
}

func is32BitRuntime() bool {
	return strconv.IntSize == 32
}
