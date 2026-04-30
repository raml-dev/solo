// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package appinfo

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type parsedVersion struct {
	major            int
	minor            int
	patch            int
	prerelease       string
	prereleasePrefix string
	prereleaseNumber int
	hasPreNumber     bool
}

var prereleasePattern = regexp.MustCompile(`^([A-Za-z]+)[\.-]?(\d+)?$`)

func parseRepository(repositoryURL string) (owner string, repo string) {
	raw := strings.TrimSpace(repositoryURL)
	if raw == "" {
		raw = defaultRepositoryURL
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "raml-dev", "solo"
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return "raml-dev", "solo"
	}

	return parts[0], strings.TrimSuffix(parts[1], ".git")
}

func normalizedReleaseTime(release GitHubRelease) time.Time {
	if !release.UpdatedAt.IsZero() {
		return release.UpdatedAt
	}
	return release.CreatedAt
}

func shouldIncludePrereleases(currentVersion string, allowConfiguredPrereleases bool) bool {
	if allowConfiguredPrereleases {
		return true
	}

	current := strings.TrimSpace(strings.TrimPrefix(currentVersion, "v"))
	if current == "" || strings.EqualFold(current, "dev") {
		return true
	}

	parsed, ok := parseVersion(current)
	return ok && parsed.prerelease != ""
}

func isCurrentVersionOlder(currentVersion, latestVersion string) bool {
	current := strings.TrimSpace(currentVersion)
	latest := strings.TrimSpace(latestVersion)

	if strings.EqualFold(current, "dev") {
		return latest != ""
	}

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
	if len(parts) != 3 {
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
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return parsedVersion{}, false
	}

	parsed := parsedVersion{
		major:      major,
		minor:      minor,
		patch:      patch,
		prerelease: prerelease,
	}

	if prerelease != "" {
		matches := prereleasePattern.FindStringSubmatch(prerelease)
		if len(matches) == 3 {
			parsed.prereleasePrefix = strings.ToLower(matches[1])
			if matches[2] != "" {
				number, numberErr := strconv.Atoi(matches[2])
				if numberErr == nil {
					parsed.prereleaseNumber = number
					parsed.hasPreNumber = true
				}
			}
		}
	}

	return parsed, true
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

	if left.prerelease == "" && right.prerelease != "" {
		return 1
	}
	if left.prerelease != "" && right.prerelease == "" {
		return -1
	}
	if left.prerelease == "" && right.prerelease == "" {
		return 0
	}

	if left.prereleasePrefix != "" && right.prereleasePrefix != "" && left.prereleasePrefix != right.prereleasePrefix {
		return strings.Compare(left.prereleasePrefix, right.prereleasePrefix)
	}
	if left.hasPreNumber && right.hasPreNumber && left.prereleasePrefix == right.prereleasePrefix {
		switch {
		case left.prereleaseNumber > right.prereleaseNumber:
			return 1
		case left.prereleaseNumber < right.prereleaseNumber:
			return -1
		default:
			return 0
		}
	}

	return strings.Compare(left.prerelease, right.prerelease)
}
