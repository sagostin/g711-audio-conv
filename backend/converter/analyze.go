package converter

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// AudioStats holds loudness analysis results from ffmpeg's loudnorm filter.
type AudioStats struct {
	InputLoudness  float64 `json:"input_loudness"`  // Integrated loudness (LUFS)
	InputTruePeak  float64 `json:"input_true_peak"` // True peak (dBTP)
	InputLRA       float64 `json:"input_lra"`       // Loudness Range
	InputThreshold float64 `json:"input_threshold"` // Loudness threshold
}

// loudnormOutput maps the JSON keys from ffmpeg's loudnorm print_format=json output.
type loudnormOutput struct {
	InputI       string `json:"input_i"`
	InputTP      string `json:"input_tp"`
	InputLRA     string `json:"input_lra"`
	InputThresh  string `json:"input_thresh"`
	OutputI      string `json:"output_i"`
	OutputTP     string `json:"output_tp"`
	OutputLRA    string `json:"output_lra"`
	OutputThresh string `json:"output_thresh"`
	NormType     string `json:"normalization_type"`
	TargetOffset string `json:"target_offset"`
}

// jsonBlockRe matches the JSON object that loudnorm prints to stderr.
var jsonBlockRe = regexp.MustCompile(`(?s)\{[^{}]*"input_i"[^{}]*\}`)

// AnalyzeAudio runs ffmpeg's loudnorm filter in measurement-only mode to extract
// loudness statistics for the given audio file.
func AnalyzeAudio(filePath string) (AudioStats, error) {
	args := []string{
		"-i", filePath,
		"-af", "loudnorm=print_format=json",
		"-f", "null",
		"-",
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return AudioStats{}, fmt.Errorf("ffmpeg analyze failed: %v (output: %s)", err, string(output))
	}

	return parseLoudnormJSON(string(output))
}

// parseLoudnormJSON extracts the loudnorm JSON block from ffmpeg output and
// parses it into AudioStats.
func parseLoudnormJSON(output string) (AudioStats, error) {
	match := jsonBlockRe.FindString(output)
	if match == "" {
		return AudioStats{}, fmt.Errorf("no loudnorm JSON found in ffmpeg output")
	}

	// Clean up any whitespace issues
	match = strings.TrimSpace(match)

	var ln loudnormOutput
	if err := json.Unmarshal([]byte(match), &ln); err != nil {
		return AudioStats{}, fmt.Errorf("failed to parse loudnorm JSON: %v", err)
	}

	stats := AudioStats{
		InputLoudness:  parseFloat(ln.InputI),
		InputTruePeak:  parseFloat(ln.InputTP),
		InputLRA:       parseFloat(ln.InputLRA),
		InputThreshold: parseFloat(ln.InputThresh),
	}

	return stats, nil
}

// BuildNormalizationFilters constructs the second-pass loudnorm filter for
// EBU R128 normalization.
//
// Pass 1 (AnalyzeAudio) measures the input's loudness statistics. This function
// builds the Pass 2 filter that feeds those measurements back into loudnorm so
// it can apply a precise linear gain adjustment to hit the target LUFS.
//
// Parameters:
//   - targetDB: desired integrated loudness in LUFS (e.g. -6 for auto attendant)
//   - stats: measured loudness from Pass 1
//   - TP=-1.5: true peak ceiling to prevent clipping
//   - LRA=11: generous loudness range allowance (won't over-compress)
//   - linear=true: single linear gain — preserves original dynamics
func BuildNormalizationFilters(stats AudioStats, targetDB float64) []string {
	return []string{
		fmt.Sprintf(
			"loudnorm=I=%.1f:TP=-1.5:LRA=11:measured_I=%.2f:measured_TP=%.2f:measured_LRA=%.2f:measured_thresh=%.2f:offset=0:linear=true:print_format=summary",
			targetDB,
			stats.InputLoudness,
			stats.InputTruePeak,
			stats.InputLRA,
			stats.InputThreshold,
		),
	}
}

// parseFloat converts a string to float64, returning 0 on error.
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// FormatStats returns a human-readable summary of the audio stats.
func (s AudioStats) FormatStats() string {
	return fmt.Sprintf("loudness: %.1f LUFS, peak: %.1f dBTP, LRA: %.1f",
		s.InputLoudness, s.InputTruePeak, s.InputLRA)
}
