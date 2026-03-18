package converter

import (
	"encoding/json"
	"fmt"
	"math"
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

// BuildNormalizationFilters constructs the filter chain for normalization:
//
//  1. loudnorm (two-pass, linear) — adjusts integrated loudness to the target
//     LUFS using measured stats from Pass 1. True peak ceiling is set to the
//     target so loudnorm itself avoids exceeding it.
//
//  2. alimiter (brick-wall limiter) — hard ceiling at the target dB. Any
//     transient peaks that still exceed the target are transparently limited.
//     Nothing in the output will be louder than the target.
func BuildNormalizationFilters(stats AudioStats, targetDB float64) []string {
	// Convert target dB to linear amplitude for alimiter (e.g. -6 dB → 0.501)
	limit := math.Pow(10, targetDB/20.0)

	return []string{
		// Step 1: Normalize integrated loudness to target LUFS
		fmt.Sprintf(
			"loudnorm=I=%.1f:TP=%.1f:LRA=11:measured_I=%.2f:measured_TP=%.2f:measured_LRA=%.2f:measured_thresh=%.2f:offset=0:linear=true:print_format=summary",
			targetDB,
			targetDB,
			stats.InputLoudness,
			stats.InputTruePeak,
			stats.InputLRA,
			stats.InputThreshold,
		),
		// Step 2: Brick-wall limiter — hard ceiling at target, nothing louder
		fmt.Sprintf("alimiter=limit=%f:level_in=1:level_out=1:attack=0.1:release=50", limit),
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
