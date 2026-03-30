package converter

import (
	"path/filepath"
	"strings"
)

// FileType represents the detected type of an audio file based on prefix.
type FileType struct {
	Prefix      string
	Label       string
	TargetDB    float64
	Description string
}

var prefixMap = map[string]FileType{
	"bicom_": {
		Prefix:      "bicom_",
		Label:       "bicom_greeting",
		TargetDB:    -12,
		Description: "Bicom Greeting",
	},
	"aa_": {
		Prefix:      "aa_",
		Label:       "auto_attendant",
		TargetDB:    -6,
		Description: "Auto Attendant",
	},
	"mbx_": {
		Prefix:      "mbx_",
		Label:       "mailbox_greeting",
		TargetDB:    -6,
		Description: "Mailbox Greeting",
	},
	"moh_": {
		Prefix:      "moh_",
		Label:       "hold_music",
		TargetDB:    -20,
		Description: "Hold Music",
	},
}

// DefaultFileType is used when no recognized prefix is found.
var DefaultFileType = FileType{
	Prefix:      "",
	Label:       "unknown",
	TargetDB:    -6,
	Description: "Unknown Type",
}

// DetectPrefix identifies the file type based on its filename prefix.
func DetectPrefix(filename string) FileType {
	base := filepath.Base(filename)
	lower := strings.ToLower(base)

	for prefix, ft := range prefixMap {
		if strings.HasPrefix(lower, prefix) {
			return ft
		}
	}

	return DefaultFileType
}

// GetAllPrefixes returns all known prefix configurations.
func GetAllPrefixes() map[string]FileType {
	result := make(map[string]FileType)
	for k, v := range prefixMap {
		result[k] = v
	}
	return result
}
