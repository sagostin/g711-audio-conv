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

// BuildNormalizationFilters constructs filters for telephony-grade normalization.
//
// The chain does two things:
//
//  1. dynaudnorm — adaptive frame-by-frame gain control that evens out the audio:
//     quiet sections are brought up, loud sections are held steady. This makes the
//     whole file sound roughly the same level (like an AGC / compressor).
//     Output peaks land near 0 dBFS.
//
//  2. volume — since dynaudnorm brings peaks to ~0 dBFS, this simple scalar
//     shifts the entire waveform down to the exact target dB ceiling.
//     Nothing can exceed the target.
//
// Parameters:
//   - p=0.95: dynaudnorm peak target (95% = -0.45 dBFS headroom)
//   - s=5: 5-second smoothing window for natural-sounding gain changes
//   - targetDB: applied as a volume scalar after dynaudnorm flattening
func BuildNormalizationFilters(stats AudioStats, targetDB float64) []string {
	return []string{
		// Step 1: Compress dynamic range — even out all levels
		"dynaudnorm=p=0.95:s=5",
		// Step 2: Scale peak to target (dynaudnorm outputs near 0 dBFS)
		fmt.Sprintf("volume=%.2fdB", targetDB),
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
