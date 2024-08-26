package update

import (
	"github.com/goccy/go-json"
	"github.com/hashicorp/go-version"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var Version = "dev"
var Commit = "none"

type GithubRelease struct {
	Url       string `json:"url"`
	AssetsUrl string `json:"assets_url"`
	UploadUrl string `json:"upload_url"`
	HtmlUrl   string `json:"html_url"`
	Id        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		Id                int    `json:"id"`
		NodeId            string `json:"node_id"`
		AvatarUrl         string `json:"avatar_url"`
		GravatarId        string `json:"gravatar_id"`
		Url               string `json:"url"`
		HtmlUrl           string `json:"html_url"`
		FollowersUrl      string `json:"followers_url"`
		FollowingUrl      string `json:"following_url"`
		GistsUrl          string `json:"gists_url"`
		StarredUrl        string `json:"starred_url"`
		SubscriptionsUrl  string `json:"subscriptions_url"`
		OrganizationsUrl  string `json:"organizations_url"`
		ReposUrl          string `json:"repos_url"`
		EventsUrl         string `json:"events_url"`
		ReceivedEventsUrl string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeId          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		Url      string `json:"url"`
		Id       int    `json:"id"`
		NodeId   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			Id                int    `json:"id"`
			NodeId            string `json:"node_id"`
			AvatarUrl         string `json:"avatar_url"`
			GravatarId        string `json:"gravatar_id"`
			Url               string `json:"url"`
			HtmlUrl           string `json:"html_url"`
			FollowersUrl      string `json:"followers_url"`
			FollowingUrl      string `json:"following_url"`
			GistsUrl          string `json:"gists_url"`
			StarredUrl        string `json:"starred_url"`
			SubscriptionsUrl  string `json:"subscriptions_url"`
			OrganizationsUrl  string `json:"organizations_url"`
			ReposUrl          string `json:"repos_url"`
			EventsUrl         string `json:"events_url"`
			ReceivedEventsUrl string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadUrl string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballUrl string `json:"tarball_url"`
	ZipballUrl string `json:"zipball_url"`
	Body       string `json:"body"`
}

func CheckForUpdates() (bool, error) {
	log.Println("Checking for updates...")
	if Version == "dev" {
		return false, nil
	}
	githubAPI := "https://api.github.com/repos/mja00/kami-chan-server-installer/releases/latest"
	// Make the request
	resp, err := http.Get(githubAPI)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	// This will be a JSON object
	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, err
	}
	repoVersion, err := version.NewVersion(release.TagName)
	currentVersion, err := version.NewVersion(Version)
	if err != nil {
		return false, err
	}
	// If the repo version is greater than the current version, then we can update
	if repoVersion.GreaterThan(currentVersion) {
		return true, nil
	}
	return false, nil
}

func DownloadUpdate(url string) error {
	// Stream the file into our temp directory
	// Create the temp directory if it doesn't exist
	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		err := os.MkdirAll("temp", 0755)
		if err != nil {
			return err
		}
	}
	out, err := os.Create(filepath.Join("temp", "update.tar.gz"))
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading update",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return err
	}

	return nil
}
