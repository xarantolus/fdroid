package apps

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/go-github/v90/github"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// FindAPKRelease picks the single APK to publish for a release, or nil if it
// has none. Only one APK per release is supported: they are stored as
// <app>_<tag>.apk, so several would overwrite each other. A release carrying
// per-ABI splits is therefore ambiguous unless one of them is the universal
// APK, and is reported rather than resolved by guessing.
func FindAPKRelease(release *github.RepositoryRelease) (*github.ReleaseAsset, error) {
	var candidates []*github.ReleaseAsset
	for _, asset := range release.Assets {
		if asset.GetState() == "uploaded" && strings.HasSuffix(asset.GetName(), ".apk") {
			candidates = append(candidates, asset)
		}
	}

	switch len(candidates) {
	case 0:
		return nil, nil
	case 1:
		return candidates[0], nil
	}

	for _, asset := range candidates {
		if strings.Contains(strings.ToLower(asset.GetName()), "universal") {
			return asset, nil
		}
	}

	names := make([]string, 0, len(candidates))
	for _, asset := range candidates {
		names = append(names, asset.GetName())
	}

	return nil, fmt.Errorf("release %q has %d APK assets and none is universal: %s",
		release.GetTagName(), len(candidates), strings.Join(names, ", "))
}

func GenerateReleaseFilename(appName string, tagName string) string {
	var normalName = fmt.Sprintf("%s_%s.apk", appName, tagName)

	var tc = transform.Chain(norm.NFD, runes.Remove(runes.Predicate(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	})), norm.NFC)

	cleaned, _, err := transform.String(tc, normalName)
	if err != nil {
		cleaned = normalName
	}

	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
		}
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		if r == '_' || r == '-' || r == '.' {
			return r
		}
		return -1
	}, cleaned)
}

func ListAllReleases(githubClient *github.Client, appRepoAuthor, appRepoName string) (allReleases []*github.RepositoryRelease, err error) {
	var currentPage int = 1

	for {
		rels, _, ierr := githubClient.Repositories.ListReleases(context.Background(), appRepoAuthor, appRepoName, &github.ListOptions{
			Page:    currentPage,
			PerPage: 100,
		})
		if ierr != nil || len(rels) == 0 {
			err = ierr
			break
		}

		allReleases = append(allReleases, rels...)
		currentPage++
	}

	return
}
