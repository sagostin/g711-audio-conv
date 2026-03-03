package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tops-audio-conv/converter"
)

const defaultMaxBulkUploadBytes = 200 * 1024 * 1024 // 200MB

// Supported audio extensions for extraction from ZIP.
var audioExtensions = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".ogg":  true,
	".flac": true,
	".m4a":  true,
	".wma":  true,
	".aac":  true,
	".gsm":  true,
	".raw":  true,
}

// BulkConvertHandler handles POST /api/convert/bulk for ZIP uploads.
func BulkConvertHandler(maxUploadMB int64) http.HandlerFunc {
	maxBytes := maxUploadMB * 1024 * 1024
	if maxBytes <= 0 {
		maxBytes = defaultMaxBulkUploadBytes
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

		if err := r.ParseMultipartForm(maxBytes); err != nil {
			http.Error(w, fmt.Sprintf("File too large or invalid form: %v", err), http.StatusBadRequest)
			return
		}
		defer r.MultipartForm.RemoveAll()

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, fmt.Sprintf("No file provided: %v", err), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Verify it's a ZIP file
		if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
			http.Error(w, "Only ZIP files are supported for bulk conversion", http.StatusBadRequest)
			return
		}

		opts := parseConvertOptions(r)

		// Create unique job directory
		jobID := generateJobID()
		jobDir, inputDir, outputDir, err := createJobDir(jobID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer os.RemoveAll(jobDir)

		log.Printf("Job %s: bulk conversion from %s", jobID, header.Filename)

		// Save uploaded ZIP
		zipPath := filepath.Join(jobDir, header.Filename)
		dst, err := os.Create(zipPath)
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

		// Extract ZIP
		audioFiles, err := extractZip(zipPath, inputDir)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to extract ZIP: %v", err), http.StatusBadRequest)
			return
		}

		if len(audioFiles) == 0 {
			http.Error(w, "No audio files found in ZIP", http.StatusBadRequest)
			return
		}

		// Convert each file and build log
		format := converter.Formats[opts.Format]
		var logEntries []string
		logEntries = append(logEntries, "=== Audio Converter — Conversion Log ===")
		logEntries = append(logEntries, fmt.Sprintf("Job ID: %s", jobID))
		logEntries = append(logEntries, fmt.Sprintf("Timestamp: %s", time.Now().Format(time.RFC3339)))
		logEntries = append(logEntries, fmt.Sprintf("Source: %s", header.Filename))
		logEntries = append(logEntries, fmt.Sprintf("Output Format: %s", format.Label))
		logEntries = append(logEntries, fmt.Sprintf("Normalization: %v", opts.Normalize))
		logEntries = append(logEntries, fmt.Sprintf("Bandpass: %v (%.0fHz - %.0fHz)", opts.Bandpass, opts.BandpassLow, opts.BandpassHigh))
		logEntries = append(logEntries, fmt.Sprintf("Total Files: %d", len(audioFiles)))
		logEntries = append(logEntries, "")
		logEntries = append(logEntries, "--- Per-File Results ---")
		logEntries = append(logEntries, "")

		successCount := 0
		failCount := 0

		for _, audioFile := range audioFiles {
			baseName := filepath.Base(audioFile)
			fileType := converter.DetectPrefix(baseName)

			// Build per-file options
			fileOpts := opts
			fileOpts.InputPath = audioFile
			nameNoExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
			outputName := nameNoExt + "_converted" + format.Extension
			fileOpts.OutputPath = filepath.Join(outputDir, outputName)

			if fileOpts.Normalize {
				// Use user-specified target_db if provided, otherwise fall back to prefix default
				if fileOpts.TargetDB == 0 {
					fileOpts.TargetDB = fileType.TargetDB
				}
			}

			// Analyze input audio
			inputStats, inputErr := converter.AnalyzeAudio(audioFile)
			if inputErr != nil {
				log.Printf("Job %s: INPUT analysis failed for %s — %v", jobID, baseName, inputErr)
			} else {
				log.Printf("Job %s: INPUT  [%s] — %s", jobID, baseName, inputStats.FormatStats())
			}

			result := converter.Convert(fileOpts)

			if result.Success {
				successCount++

				// Analyze output audio
				outputStats, outputErr := converter.AnalyzeAudio(fileOpts.OutputPath)
				if outputErr != nil {
					log.Printf("Job %s: OUTPUT analysis failed for %s — %v", jobID, baseName, outputErr)
				} else {
					log.Printf("Job %s: OUTPUT [%s] — %s", jobID, baseName, outputStats.FormatStats())
				}

				logEntries = append(logEntries, fmt.Sprintf("[OK]  %s", baseName))
				logEntries = append(logEntries, fmt.Sprintf("      Type: %s | Target: %.1f dB | Output: %s",
					fileType.Description, fileOpts.TargetDB, outputName))
				if inputErr == nil {
					logEntries = append(logEntries, fmt.Sprintf("      INPUT  — %s", inputStats.FormatStats()))
				}
				if outputErr == nil {
					logEntries = append(logEntries, fmt.Sprintf("      OUTPUT — %s", outputStats.FormatStats()))
				}
			} else {
				failCount++
				logEntries = append(logEntries, fmt.Sprintf("[FAIL] %s", baseName))
				logEntries = append(logEntries, fmt.Sprintf("       Error: %s", result.Error))
			}
			logEntries = append(logEntries, "")
		}

		logEntries = append(logEntries, "--- Summary ---")
		logEntries = append(logEntries, fmt.Sprintf("Success: %d | Failed: %d | Total: %d", successCount, failCount, len(audioFiles)))

		// Write conversion log
		logPath := filepath.Join(outputDir, "conversion_log.txt")
		os.WriteFile(logPath, []byte(strings.Join(logEntries, "\n")), 0644)

		log.Printf("Job %s: bulk complete — %d success, %d failed", jobID, successCount, failCount)

		// Create output ZIP
		outputZipName := strings.TrimSuffix(header.Filename, ".zip") + "_converted.zip"
		outputZipPath := filepath.Join(jobDir, outputZipName)
		if err := createZip(outputDir, outputZipPath); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create output ZIP: %v", err), http.StatusInternalServerError)
			return
		}

		// Send ZIP as download
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, outputZipName))
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("X-Job-ID", jobID)

		outputZip, err := os.Open(outputZipPath)
		if err != nil {
			http.Error(w, "Failed to read output ZIP", http.StatusInternalServerError)
			return
		}
		defer outputZip.Close()

		stat, _ := outputZip.Stat()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
		io.Copy(w, outputZip)
	}
}

// extractZip extracts audio files from a ZIP archive to the target directory.
func extractZip(zipPath, targetDir string) ([]string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var audioFiles []string

	for _, f := range reader.File {
		// Skip directories and hidden files
		if f.FileInfo().IsDir() || strings.HasPrefix(filepath.Base(f.Name), ".") {
			continue
		}

		// Check if it's an audio file
		ext := strings.ToLower(filepath.Ext(f.Name))
		if !audioExtensions[ext] {
			continue
		}

		// Extract file (flatten directory structure)
		baseName := filepath.Base(f.Name)
		destPath := filepath.Join(targetDir, baseName)

		// Handle filename conflicts
		if _, err := os.Stat(destPath); err == nil {
			nameNoExt := strings.TrimSuffix(baseName, ext)
			destPath = filepath.Join(targetDir, nameNoExt+"_dup"+ext)
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			continue
		}

		io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		audioFiles = append(audioFiles, destPath)
	}

	return audioFiles, nil
}

// createZip creates a ZIP archive from all files in a directory.
func createZip(sourceDir, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer w.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		writer, err := w.Create(relPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}
