package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// ImageManifestList represents the structure of a multi-arch image manifest.
type ImageManifestList struct {
	Manifests []struct {
		Digest    string          `json:"digest"`
		MediaType types.MediaType `json:"mediaType"`
		Platform  struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
		} `json:"platform"`
	} `json:"manifests"`
}

func getChecksum() {}

func deleteSignature() {}

func main() {
	ghToken := os.Getenv("GH_TOKEN")
	ghOrg := os.Getenv("GH_ORG")
	ghUser := os.Getenv("GH_USER")
	packageName := os.Getenv("PACKAGE_NAME")
	packageType := os.Getenv("PACKAGE_TYPE")
	versionTag := os.Getenv("TAG")
	dryrun_env := os.Getenv("DRYRUN")

	if ghToken == "" {
		slog.Error("Missing required environment variable: GH_TOKEN")
		os.Exit(1)
	}

	if packageName == "" {
		slog.Error("Missing required environment variable: PACKAGE_NAME")
		os.Exit(1)
	}

	if versionTag == "" {
		slog.Error("Missing required environment variable: TAG")
		os.Exit(1)
	}

	if ghOrg != "" && ghUser != "" {
		slog.Error("Please only provide either a github user or a github org.")
		os.Exit(1)
	}

	username := ""

	if ghOrg != "" {
		username = ghOrg
	} else {
		username = ghUser
	}

	if dryrun_env == "" {
		dryrun_env = "false"
	}

	dryrun, err := strconv.ParseBool(dryrun_env)
	if err != nil {
		slog.Error("Error parsing bool", "Error", err)
	}

	if packageType == "" {
		packageType = "container"
	}

	auth := authn.FromConfig(authn.AuthConfig{
		Username: username,
		Password: ghToken,
	})

	packageURL := fmt.Sprintf("ghcr.io/%s/%s:%s", username, packageName, versionTag)

	slog.Info("Fetching image metadata...")

	manifestData, err := crane.Manifest(packageURL, crane.WithAuth(auth))
	if err != nil {
		slog.Error("Error fetching manifest.", "Error", err)
		os.Exit(1)
	}

	var manifestList ImageManifestList
	if err := json.Unmarshal(manifestData, &manifestList); err != nil {
		slog.Error("Error parsing manifest", "Error", err)
		os.Exit(1)
	}

	// Append to GitHub summary
	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Could not open summary file", "Error", err)
	}
	defer f.Close()

	imgTagged := false
	taggedVersions := ""

	if imgTagged {
		if dryrun {
			f.WriteString(":warning: This is a dry run, no versions were actually tagged.\n\n")
		}

		f.WriteString("## Pruned Cosign Signatures\n\n")
		f.WriteString("| Tags |\n|--------------|\n")
		f.WriteString(taggedVersions + "\n")
	} else {
		f.WriteString("No orphaned signatures found.")
	}
}
