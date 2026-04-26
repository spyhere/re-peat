package autoupdate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	releasesUrl = "https://api.github.com/repos/spyhere/re-peat/releases"
	appName     = "re-peat"
)

type releaseDTO struct {
	HtmlUrl     string    `json:"html_url"`
	TagName     tagName   `json:"tag_name"`
	Name        string    `json:"name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []asset   `json:"assets"`
	Body        string    `json:"body"`
}

type release struct {
	HtmlUrl     string
	TagName     tagName
	Name        string
	PublishedAt time.Time
	Asset       asset
	Body        string
}

type asset struct {
	Name               string `json:"name"`
	Size               int    `json:"size"`
	Digest             string `json:"digest"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

func getLatestRelease() (releaseDTO, error) {
	res, err := http.Get(releasesUrl)
	if err != nil {
		return releaseDTO{}, err
	}
	defer res.Body.Close()

	var releases []releaseDTO
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&releases); err != nil {
		return releaseDTO{}, err
	}
	for _, it := range releases {
		if it.Draft || it.Prerelease {
			continue
		}
		return it, nil
	}
	return releaseDTO{}, nil
}

func isSameDay(checkTime time.Time) bool {
	if checkTime.IsZero() {
		return false
	}
	curY, curM, curD := time.Now().UTC().Date()
	checkY, checkM, checkD := checkTime.Date()
	return curY == checkY && curM == checkM && curD == checkD
}

func ShouldUpdate(tag string, lastCheckDate time.Time) (release, bool, error) {
	if tag == "dev" || isSameDay(lastCheckDate) {
		return release{}, false, nil
	}
	rel, err := getLatestRelease()
	if err != nil {
		return release{}, false, err
	}
	if rel.TagName.isLessOrEqual(tag) {
		return release{}, false, nil
	}
	if len(rel.Assets) == 0 {
		return release{}, false, nil
	}
	prefix := fmt.Sprintf("%s_%s", appName, runtime.GOOS)
	var validAsset asset
	for _, it := range rel.Assets {
		if strings.HasPrefix(it.Name, prefix) {
			validAsset = it
			break
		}
	}
	if validAsset.BrowserDownloadUrl == "" {
		return release{}, false, nil
	}
	return release{
		HtmlUrl:     rel.HtmlUrl,
		TagName:     rel.TagName,
		Name:        rel.Name,
		PublishedAt: rel.PublishedAt,
		Asset:       validAsset,
		Body:        rel.Body,
	}, true, nil
}
