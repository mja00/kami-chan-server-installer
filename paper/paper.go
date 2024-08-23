package paper

import (
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// This will handle all of our API calls to the Paper API

const baseURL = "https://papermc.io/api/v2"

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

func (p *PaperAPI) GetProjects() ([]string, error) {
	req, err := http.NewRequest("GET", baseURL+"/projects", nil)
	if err != nil {
		return nil, err
	}

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

func (p *PaperAPI) GetBuild(projectID, version string, build int) (*BuildResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/projects/"+projectID+"/versions/"+version+"/builds/"+strconv.Itoa(build), nil)
	if err != nil {
		return nil, err
	}

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

func (p *PaperAPI) GetBuildDownload(projectID, version string, build int, download string, outputPath string) error {
	// This will download the file to the given path
	req, err := http.NewRequest("GET", baseURL+"/projects/"+projectID+"/versions/"+version+"/builds/"+strconv.Itoa(build)+"/downloads/"+download, nil)
	if err != nil {
		return err
	}

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

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
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

	err = p.GetBuildDownload(projectID, version, build.Build, build.Downloads.Application.Name, outputPath)
	if err != nil {
		return err
	}

	return nil
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
	return p.GetBuildDownload(projectID, version, build.Build, build.Downloads.Application.Name, outputPath)
}
