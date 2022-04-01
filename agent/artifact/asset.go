package artifact

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type asset struct {
	FullName    string
	DownloadURL string
}

func (a *asset) stripExtension() string {
	return strings.TrimSuffix(a.FullName, filepath.Ext(a.FullName))
}

func (a *asset) Name() string {
	stripped := a.stripExtension()
	if stripped == "" {
		return "unknown"
	}

	return strings.Split(stripped, "-")[0]
}

func (a *asset) Architecture() string {
	var supportedArchs = []string{"amd64", "arm64", "arm"}

	for _, arch := range supportedArchs {
		for _, alias := range architectureAliases(arch) {
			if strings.Contains(a.FullName, fmt.Sprintf("-%s", alias)) {
				return arch
			}
		}
	}
	return "unknown"
}

func (a *asset) Platform() string {
	var supportedPlatforms = []string{"linux", "darwin"}

	for _, platform := range supportedPlatforms {
		for _, alias := range platformAliases(platform) {
			matches, err := regexp.Match(fmt.Sprintf("-%s[-\\.]", alias), []byte(a.FullName))
			if err != nil {
				continue
			}
			if matches {
				return platform
			}
		}
	}

	return "unknown"
}
