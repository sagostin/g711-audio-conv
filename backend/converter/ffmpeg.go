package converter

import (
	"fmt"
	"os/exec"
	"strings"
)

// OutputFormat defines a supported audio output format.
type OutputFormat struct {
	ID         string
	Label      string
	Extension  string
	SampleRate int
	Channels   int
	Codec      string
	ExtraArgs  []string
}

// Supported output formats.
var Formats = map[string]OutputFormat{
	"wav-pcm": {
		ID:         "wav-pcm",
		Label:      "Standard WAV (8kHz, 16-bit PCM)",
		Extension:  ".wav",
		SampleRate: 8000,
		Channels:   1,
		Codec:      "pcm_s16le",
	},
	"wav-ulaw": {
		ID:         "wav-ulaw",
		Label:      "G.711 µ-law WAV (8kHz)",
		Extension:  ".wav",
		SampleRate: 8000,
		Channels:   1,
		Codec:      "pcm_mulaw",
	},
	"wav-alaw": {
		ID:         "wav-alaw",
		Label:      "G.711 A-law WAV (8kHz)",
		Extension:  ".wav",
		SampleRate: 8000,
		Channels:   1,
		Codec:      "pcm_alaw",
	},
	"g722": {
		ID:         "g722",
		Label:      "G.722 (16kHz)",
		Extension:  ".wav",
		SampleRate: 16000,
		Channels:   1,
		Codec:      "g722",
	},
}

// ConvertOptions holds all conversion parameters.
type ConvertOptions struct {
	InputPath    string
	OutputPath   string
	Format       string // format key from Formats map
	Normalize    bool
	TargetDB     float64 // normalization target in dB
	Bandpass     bool
	BandpassLow  float64 // highpass cutoff Hz
	BandpassHigh float64 // lowpass cutoff Hz
}

// ConvertResult holds the result of a conversion.
type ConvertResult struct {
	Success      bool
	OutputPath   string
	Format       string
	TargetDB     float64
	Bandpass     bool
	BandpassLow  float64
	BandpassHigh float64
	Error        string
	FFmpegCmd    string
	InputStats   AudioStats
	OutputStats  AudioStats
}

// Convert runs ffmpeg to convert an audio file with the given options.
// When normalization is enabled, it uses a two-pass approach:
//
//	Pass 1: AnalyzeAudio() measures input loudness stats
//	Pass 2: Feeds measured values into loudnorm with linear=true for precise normalization
//
// This avoids single-pass dynamic mode which can alter the audio's dynamic range.
func Convert(opts ConvertOptions) ConvertResult {
	result := ConvertResult{
		Format:       opts.Format,
		TargetDB:     opts.TargetDB,
		Bandpass:     opts.Bandpass,
		BandpassLow:  opts.BandpassLow,
		BandpassHigh: opts.BandpassHigh,
	}

	format, ok := Formats[opts.Format]
	if !ok {
		format = Formats["wav-pcm"]
		result.Format = "wav-pcm"
	}

	// Build filter chain
	var filters []string

	// Bandpass filter (highpass + lowpass)
	if opts.Bandpass {
		low := opts.BandpassLow
		high := opts.BandpassHigh
		if low <= 0 {
			low = 300 // telephony default
		}
		if high <= 0 {
			high = 3400 // telephony default
		}
		result.BandpassLow = low
		result.BandpassHigh = high
		filters = append(filters, fmt.Sprintf("highpass=f=%.0f", low))
		filters = append(filters, fmt.Sprintf("lowpass=f=%.0f", high))
	}

	// Linear normalization via volume + hard limiter
	if opts.Normalize {
		// Pass 1: Measure input loudness statistics
		inputStats, err := AnalyzeAudio(opts.InputPath)
		if err != nil {
			// Fall back to simple loudnorm if measurement fails
			filters = append(filters, fmt.Sprintf("loudnorm=I=%.1f:TP=-1.5:LRA=11", opts.TargetDB))
		} else {
			result.InputStats = inputStats
			// Apply exact linear gain + hard limiter (no LRA tolerance)
			filters = append(filters, BuildNormalizationFilters(inputStats, opts.TargetDB)...)
		}
	}

	// Build ffmpeg command
	args := []string{
		"-y", // overwrite output
		"-i", opts.InputPath,
	}

	// Add filter chain
	if len(filters) > 0 {
		args = append(args, "-af", strings.Join(filters, ","))
	}

	// Add codec and format options
	args = append(args,
		"-ar", fmt.Sprintf("%d", format.SampleRate),
		"-ac", fmt.Sprintf("%d", format.Channels),
		"-acodec", format.Codec,
	)

	// Add any extra args for the format
	args = append(args, format.ExtraArgs...)

	// Output path
	args = append(args, opts.OutputPath)

	result.FFmpegCmd = "ffmpeg " + strings.Join(args, " ")

	// Execute ffmpeg (pass 2 when normalizing, or single pass otherwise)
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("ffmpeg error: %v\nOutput: %s", err, string(output))
		return result
	}

	result.Success = true
	result.OutputPath = opts.OutputPath
	return result
}

// GetFormatList returns all available formats for the API.
func GetFormatList() []OutputFormat {
	list := make([]OutputFormat, 0, len(Formats))
	for _, f := range Formats {
		list = append(list, f)
	}
	return list
}
