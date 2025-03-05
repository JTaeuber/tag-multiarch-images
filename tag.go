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
		Annotations map[string]string `json:"annotations"`
	} `json:"manifests"`
}

func findMatchingDigest(ref string, list ImageManifestList) (string, string) {
	for _, image := range list.Manifests {
		if ref == image.Digest {
			return image.Platform.Architecture, image.Platform.OS
		}
	}
	return "", ""
}

func main() {
	ghToken := os.Getenv("GH_TOKEN")
	ghUser := os.Getenv("GH_USER")
	packageName := os.Getenv("PACKAGE_NAME")
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

	if dryrun_env == "" {
		dryrun_env = "false"
	}

	dryrun, err := strconv.ParseBool(dryrun_env)
	if err != nil {
		slog.Error("Error parsing bool", "Error", err)
	}

	auth := authn.FromConfig(authn.AuthConfig{
		Username: ghUser,
		Password: ghToken,
	})

	packageURL := fmt.Sprintf("ghcr.io/%s/%s:%s", ghUser, packageName, versionTag)

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

	imgTagged := false
	taggedVersions := ""

	for _, image := range manifestList.Manifests {
		url := fmt.Sprintf("ghcr.io/%s/%s@%s", ghUser, packageName, image.Digest)

		if image.Platform.Architecture == "unknown" {
			refArch, refOs := findMatchingDigest(image.Annotations["vnd.docker.reference.digest"], manifestList)
			newTag := fmt.Sprintf("%s-attestation-manifest-%s-%s", versionTag, refOs, refArch)
			newUrl := fmt.Sprintf("ghcr.io/%s/%s:%s", ghUser, packageName, newTag)

			slog.Info("Tagging version...", "Version", image.Digest)
			if !dryrun {
				crane.Copy(url, newUrl, crane.WithAuth(auth))
			}

			imgTagged = true
			taggedVersions += fmt.Sprintf("| %s | %s |\n", image.Digest, newTag)
			continue
		}

		newTag := fmt.Sprintf("%s-%s-%s", versionTag, image.Platform.OS, image.Platform.Architecture)
		newUrl := fmt.Sprintf("ghcr.io/%s/%s:%s", ghUser, packageName, newTag)

		slog.Info("Tagging version...", "Version", image.Digest)
		if !dryrun {
			crane.Copy(url, newUrl, crane.WithAuth(auth))
		}
		imgTagged = true
		taggedVersions += fmt.Sprintf("| %s | %s |\n", image.Digest, newTag)
	}

	// Append to GitHub summary
	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Could not open summary file", "Error", err)
	}
	defer f.Close()

	f.WriteString("## Tagged versions\n\n")
	if dryrun {
		f.WriteString(":warning: This is a dry run, no versions were actually tagged.\n\n")
	}

	if imgTagged {
		f.WriteString("| Digest | New Tag |\n|--------------|--------------|\n")
		f.WriteString(taggedVersions + "\n")
	} else {
		f.WriteString("No versions to tag found.")
	}
}
