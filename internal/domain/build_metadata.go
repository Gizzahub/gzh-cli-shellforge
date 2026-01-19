package domain

import (
	"encoding/json"
	"time"
)

// BuildMetadata contains information about a build for use by deploy.
type BuildMetadata struct {
	Shell       string          `json:"shell"`
	OS          string          `json:"os"`
	GeneratedAt time.Time       `json:"generated_at"`
	Files       []BuildFileInfo `json:"files"`
}

// BuildFileInfo maps a build file to its deployment target.
type BuildFileInfo struct {
	// Source is the relative path within the build directory
	Source string `json:"source"`
	// Target is the target name (e.g., "zshrc", "config")
	Target string `json:"target"`
	// DestPath is the relative path from home directory (e.g., ".zshrc", ".config/fish/config.fish")
	DestPath string `json:"dest_path"`
}

// MetadataFileName is the name of the metadata file in the build directory.
const MetadataFileName = ".shellforge-build.json"

// ToJSON serializes metadata to JSON.
func (m *BuildMetadata) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// ParseBuildMetadata deserializes metadata from JSON.
func ParseBuildMetadata(data []byte) (*BuildMetadata, error) {
	var meta BuildMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
