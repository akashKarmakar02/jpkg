package downloader

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"jpkg/config"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

type Asset struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	DonwloadUrl string `json:"browser_download_url"`
}

type GithubRes struct {
	Assets []Asset `json:"assets"`
}

func downloadFile(name, url, dest string) error {
	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size := resp.ContentLength

	// Create a progress bar
	bar := progressbar.DefaultBytes(size,
		name,
	)

	// Start downloading the file
	resp, err = http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create a writer to update the progress bar
	writer := io.MultiWriter(out, bar)

	// Copy the content to the file while updating the progress bar
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("\nDownload complete")
	return nil
}

func HandleMavenURL(url string, libDir string) error {
	// Remove the "pkg:maven/" prefix
	trimmedURL := strings.TrimPrefix(url, "pkg:maven/")
	parts := strings.Split(trimmedURL, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid Maven URL format")
	}

	groupID := strings.ReplaceAll(parts[0], ".", "/")
	artifactVersionStr := strings.Split(parts[1], "@")
	artifactID := artifactVersionStr[0]
	version := artifactVersionStr[1]
	jarFileName := fmt.Sprintf("%s-%s.jar", artifactID, version)
	downloadURL := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s", groupID, artifactID, version, jarFileName)

	// Download the JAR file
	if _, err := os.Stat(libDir); os.IsNotExist(err) {
		os.Mkdir(libDir, os.ModePerm)
	}
	dest := filepath.Join(libDir, jarFileName)

	if err := config.SaveDependency(fmt.Sprintf("%s/%s", groupID, artifactID), "maven", version); err != nil {
		return err
	}

	return downloadFile(artifactID, downloadURL, dest)
}

// Function to handle GitHub URL
func HandleGitHubURL(url, libDir string) error {
	// Example: https://github.com/user/repo/releases/latest/download/file.jar
	parts := strings.Split(url, "/")
	if len(parts) < 5 || !strings.HasPrefix(url, "https://github.com") {
		return fmt.Errorf("invalid GitHub URL format")
	}

	// Construct the latest release API URL
	user := parts[3]
	repo := parts[4]
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", user, repo)

	// Fetch the latest release info
	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var data GithubRes
	downloadUrl := ""

	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	for _, asset := range data.Assets {
		if asset.ContentType == "application/java-archive" {
			downloadUrl = asset.DonwloadUrl
		}
	}

	if downloadUrl == "" {
		return errors.New("no jar file available in repo")
	}

	jarDownloadURL := downloadUrl
	jarFileName := filepath.Base(jarDownloadURL)

	if _, err := os.Stat(libDir); os.IsNotExist(err) {
		os.Mkdir(libDir, os.ModePerm)
	}
	dest := filepath.Join(libDir, jarFileName)

	if err := config.SaveDependency(fmt.Sprintf("%s/%s", user, repo), "github", ""); err != nil {
		return err
	}
	return downloadFile(repo, jarDownloadURL, dest)
}
