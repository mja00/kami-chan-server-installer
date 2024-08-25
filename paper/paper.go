package paper

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/mja00/kami-chan-server-installer/utils"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// This will handle all of our API calls to the Paper API

var Version = "dev"
var Commit = "none"

const baseURL = "https://api.papermc.io/v2"

type PaperAPI struct {
	client *http.Client
}

func NewPaperAPI() *PaperAPI {
	return &PaperAPI{
		client: &http.Client{},
	}
}

type ProjectsResponse struct {
	Projects []string `json:"projects"`
}

func AddHeaders(req *http.Request) {
	req.Header.Add("User-Agent", "Kami Chan Server Installer"+"/"+Version+"/"+Commit)
}

func (p *PaperAPI) GetProjects() ([]string, error) {
	req, err := http.NewRequest("GET", baseURL+"/projects", nil)
	if err != nil {
		return nil, err
	}
	AddHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var projectsResponse ProjectsResponse
	if err := json.NewDecoder(resp.Body).Decode(&projectsResponse); err != nil {
		return nil, err
	}

	return projectsResponse.Projects, nil
}

type ProjectResponse struct {
	ID            string   `json:"project_id"`
	Name          string   `json:"project_name"`
	VersionGroups []string `json:"version_groups"`
	Versions      []string `json:"versions"`
}

func (p *PaperAPI) GetProject(projectID string) (*ProjectResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/projects/"+projectID, nil)
	if err != nil {
		return nil, err
	}
	AddHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var projectResponse ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&projectResponse); err != nil {
		return nil, err
	}

	return &projectResponse, nil
}

type VersionResponse struct {
	ProjectId   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	Version     string `json:"version"`
	Builds      []int  `json:"builds"`
}

func (p *PaperAPI) GetVersion(projectID, version string) (*VersionResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/projects/"+projectID+"/versions/"+version, nil)
	if err != nil {
		return nil, err
	}
	AddHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var versionResponse VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&versionResponse); err != nil {
		return nil, err
	}

	return &versionResponse, nil
}

type BuildsResponse struct {
	ProjectId   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	Version     string `json:"version"`
	Builds      []struct {
		Build    int       `json:"build"`
		Time     time.Time `json:"time"`
		Channel  string    `json:"channel"`
		Promoted bool      `json:"promoted"`
		Changes  []struct {
			Commit  string `json:"commit"`
			Summary string `json:"summary"`
			Message string `json:"message"`
		} `json:"changes"`
		Downloads struct {
			Application struct {
				Name   string `json:"name"`
				Sha256 string `json:"sha256"`
			} `json:"application"`
		} `json:"downloads"`
	} `json:"builds"`
}

func (p *PaperAPI) GetBuilds(projectID, version string) (*BuildsResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/projects/"+projectID+"/versions/"+version+"/builds", nil)
	if err != nil {
		return nil, err
	}
	AddHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buildsResponse BuildsResponse
	if err := json.NewDecoder(resp.Body).Decode(&buildsResponse); err != nil {
		return nil, err
	}

	return &buildsResponse, nil
}

type BuildResponse struct {
	ProjectId   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Version     string    `json:"version"`
	Build       int       `json:"build"`
	Time        time.Time `json:"time"`
	Channel     string    `json:"channel"`
	Promoted    bool      `json:"promoted"`
	Changes     []struct {
		Commit  string `json:"commit"`
		Summary string `json:"summary"`
		Message string `json:"message"`
	} `json:"changes"`
	Downloads struct {
		Application struct {
			Name   string `json:"name"`
			Sha256 string `json:"sha256"`
		} `json:"application"`
	} `json:"downloads"`
}

type BuildInfo struct {
	Build    int       `json:"build"`
	Time     time.Time `json:"time"`
	Channel  string    `json:"channel"`
	Promoted bool      `json:"promoted"`
	Changes  []struct {
		Commit  string `json:"commit"`
		Summary string `json:"summary"`
		Message string `json:"message"`
	} `json:"changes"`
	Downloads struct {
		Application struct {
			Name   string `json:"name"`
			Sha256 string `json:"sha256"`
		} `json:"application"`
	} `json:"downloads"`
}

func (p *PaperAPI) GetBuild(projectID, version string, build int) (*BuildResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/projects/"+projectID+"/versions/"+version+"/builds/"+strconv.Itoa(build), nil)
	if err != nil {
		return nil, err
	}
	AddHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buildResponse BuildResponse
	if err := json.NewDecoder(resp.Body).Decode(&buildResponse); err != nil {
		return nil, err
	}

	return &buildResponse, nil
}

func (p *PaperAPI) GetLatestBuild(projectID, version string) (BuildInfo, error) {
	builds, err := p.GetBuilds(projectID, version)
	if err != nil {
		return BuildInfo{}, err
	}

	return builds.Builds[len(builds.Builds)-1], nil
}

func (p *PaperAPI) GetBuildDownload(projectID, version string, build int, download string, outputPath string) error {
	// Be smart here and check if the file already exists, if it does then we can check the sha256 hash against the latest build for this version
	// If the sha256 hash matches, then we can just return nil
	// If it doesn't match, then we need to download the file
	if _, err := os.Stat(outputPath); err == nil {
		// The file must exist, so we need to check the sha256 hash
		// First, we need to get the latest build for this version
		latestBuild, err := p.GetLatestBuild(projectID, version)
		if err != nil {
			return err
		}
		latestBuildSHA256 := latestBuild.Downloads.Application.Sha256
		// Now we need to calculate the sha256 hash of the file
		// Stream the file, that way we don't have to load the whole thing into memory
		fileHash, err := utils.GetSha256Hash(outputPath)
		if err != nil {
			return fmt.Errorf("error calculating sha256 hash: %s", err)
		}
		// If they match, then we can just return nil
		if fileHash == latestBuildSHA256 {
			log.Println("File already exists and SHA256 hash matches, skipping download")
			return nil
		}
	}
	// This will download the file to the given path
	req, err := http.NewRequest("GET", baseURL+"/projects/"+projectID+"/versions/"+version+"/builds/"+strconv.Itoa(build)+"/downloads/"+download, nil)
	if err != nil {
		return err
	}
	AddHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading Paper",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (p *PaperAPI) DownloadLatestBuild(projectID, outputPath string, allowExperimentalBuilds bool) error {
	project, err := p.GetProject(projectID)
	if err != nil {
		return err
	}
	// Need the last item in the versions list
	version := project.Versions[len(project.Versions)-1]

	builds, err := p.GetBuilds(projectID, version)
	if err != nil {
		return err
	}

	// Recursively check the builds, from last to first, until we find a build that has a channel of "default"
	// If allowExperimentalBuilds is true, we can ignore this check and just return the last build
	var build BuildInfo
	for i := len(builds.Builds) - 1; i >= 0; i-- {
		currentBuild := builds.Builds[i]
		if currentBuild.Channel == "default" || allowExperimentalBuilds {
			build = currentBuild
			break
		}
	}

	// If we don't have a build, then they were all experimental builds (probably a new MC release
	if build.Build == 0 {
		return fmt.Errorf("no builds found for %s", version)
	}

	return p.GetBuildDownload(projectID, version, build.Build, build.Downloads.Application.Name, outputPath)
}

func (p *PaperAPI) DownloadLatestBuildForVersion(projectID, version, outputPath string, allowExperimentalBuilds bool) error {
	builds, err := p.GetBuilds(projectID, version)
	if err != nil {
		return err
	}

	// Recursively check the builds, from last to first, until we find a build that has a channel of "default"
	// If allowExperimentalBuilds is true, we can ignore this check and just return the last build
	var build BuildInfo
	for i := len(builds.Builds) - 1; i >= 0; i-- {
		currentBuild := builds.Builds[i]
		if currentBuild.Channel == "default" || allowExperimentalBuilds {
			build = currentBuild
			break
		}
	}

	if build.Build == 0 {
		return fmt.Errorf("no builds found for %s", version)
	}

	return p.GetBuildDownload(projectID, version, build.Build, build.Downloads.Application.Name, outputPath)
}
