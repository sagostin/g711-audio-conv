package handlers

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"tops-audio-conv/converter"
)

const defaultMaxUploadBytes = 50 * 1024 * 1024 // 50MB

// ConversionsDir is the base directory for storing conversion jobs.
// Each job gets a unique subdirectory with input/ and output/ folders.
var ConversionsDir = getConversionsDir()

func getConversionsDir() string {
	dir := os.Getenv("CONVERSIONS_DIR")
	if dir == "" {
		dir = "./conversions"
	}
	return dir
}

// generateJobID creates a short unique identifier for a conversion job.
func generateJobID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// createJobDir creates a structured job directory: conversions/<jobID>/{input,output}
func createJobDir(jobID string) (jobDir, inputDir, outputDir string, err error) {
	jobDir = filepath.Join(ConversionsDir, jobID)
	inputDir = filepath.Join(jobDir, "input")
	outputDir = filepath.Join(jobDir, "output")

	if err = os.MkdirAll(inputDir, 0755); err != nil {
		return
	}
	if err = os.MkdirAll(outputDir, 0755); err != nil {
		return
	}
	return
}

// ConvertHandler handles POST /api/convert for single file conversion.
func ConvertHandler(maxUploadMB int64) http.HandlerFunc {
	maxBytes := maxUploadMB * 1024 * 1024
	if maxBytes <= 0 {
		maxBytes = defaultMaxUploadBytes
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

		// Parse multipart form
		if err := r.ParseMultipartForm(maxBytes); err != nil {
			http.Error(w, fmt.Sprintf("File too large or invalid form: %v", err), http.StatusBadRequest)
			return
		}
		defer r.MultipartForm.RemoveAll()

		// Get uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, fmt.Sprintf("No file provided: %v", err), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Parse conversion options from form fields
		opts := parseConvertOptions(r)

		// Detect file prefix for normalization
		fileType := converter.DetectPrefix(header.Filename)
		if opts.Normalize {
			opts.TargetDB = fileType.TargetDB
		}

		// Create unique job directory
		jobID := generateJobID()
		jobDir, inputDir, outputDir, err := createJobDir(jobID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer os.RemoveAll(jobDir) // Clean up after response is sent

		log.Printf("Job %s: converting %s (type=%s, format=%s)", jobID, header.Filename, fileType.Label, opts.Format)

		// Save uploaded file to input dir
		inputPath := filepath.Join(inputDir, header.Filename)
		dst, err := os.Create(inputPath)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(dst, file); err != nil {
			dst.Close()
			http.Error(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
		dst.Close()

		// Determine output filename
		format := converter.Formats[opts.Format]
		baseName := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
		outputName := baseName + "_converted" + format.Extension
		outputPath := filepath.Join(outputDir, outputName)

		// Set conversion paths
		opts.InputPath = inputPath
		opts.OutputPath = outputPath

		// Run conversion
		result := converter.Convert(opts)
		if !result.Success {
			log.Printf("Job %s: FAILED — %s", jobID, result.Error)
			http.Error(w, fmt.Sprintf("Conversion failed: %s", result.Error), http.StatusInternalServerError)
			return
		}

		log.Printf("Job %s: SUCCESS — %s", jobID, outputName)

		// Send converted file as download
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, outputName))
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("X-Job-ID", jobID)
		w.Header().Set("X-File-Type", fileType.Label)
		w.Header().Set("X-Normalization-DB", fmt.Sprintf("%.1f", fileType.TargetDB))

		convertedFile, err := os.Open(outputPath)
		if err != nil {
			http.Error(w, "Failed to read converted file", http.StatusInternalServerError)
			return
		}
		defer convertedFile.Close()

		stat, _ := convertedFile.Stat()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
		io.Copy(w, convertedFile)
	}
}

// FormatsHandler returns available output formats as JSON.
func FormatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	formats := converter.GetFormatList()

	// Simple JSON response
	fmt.Fprint(w, "[")
	for i, f := range formats {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `{"id":"%s","label":"%s","extension":"%s","sampleRate":%d}`,
			f.ID, f.Label, f.Extension, f.SampleRate)
	}
	fmt.Fprint(w, "]")
}

// PrefixesHandler returns known file prefixes and their normalization targets.
func PrefixesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	prefixes := converter.GetAllPrefixes()

	fmt.Fprint(w, "[")
	i := 0
	for _, p := range prefixes {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `{"prefix":"%s","label":"%s","targetDb":%.1f,"description":"%s"}`,
			p.Prefix, p.Label, p.TargetDB, p.Description)
		i++
	}
	fmt.Fprint(w, "]")
}

// parseConvertOptions extracts conversion options from form values.
func parseConvertOptions(r *http.Request) converter.ConvertOptions {
	opts := converter.ConvertOptions{
		Format:    r.FormValue("format"),
		Normalize: r.FormValue("normalize") != "false",
		Bandpass:  r.FormValue("bandpass") == "true",
	}

	if opts.Format == "" {
		opts.Format = "wav-pcm"
	}

	if _, ok := converter.Formats[opts.Format]; !ok {
		opts.Format = "wav-pcm"
	}

	if low, err := strconv.ParseFloat(r.FormValue("bandpass_low"), 64); err == nil {
		opts.BandpassLow = low
	}
	if high, err := strconv.ParseFloat(r.FormValue("bandpass_high"), 64); err == nil {
		opts.BandpassHigh = high
	}

	return opts
}
