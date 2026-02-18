# Audio Converter — Task List

## Backend (Go)
- [x] Initialize Go module + project structure
- [x] File prefix detection (`aa_`, `mbx_`, `moh_`) with normalization levels (-6dB / -20dB)
- [x] ffmpeg conversion wrapper (8kHz WAV default + optional µ-law, A-law, G.722)
- [x] Audio normalization via ffmpeg loudnorm filter
- [x] Bandpass filter via ffmpeg highpass + lowpass (300Hz–3400Hz telephony default)
- [x] Single file upload + conversion endpoint (`POST /api/convert`)
- [x] Bulk ZIP upload + convert + re-zip with `conversion_log.txt` (`POST /api/convert/bulk`)
- [x] Proxy IP forwarding middleware (env `PROXY_HEADER`)
- [x] CORS + API config
- [x] Unique job directories under `conversions/<jobID>/` with input/output structure
- [x] Easter egg mode — IP-based Comic Sans + rainbow text (`EASTER_EGG_IPS` env)

## Frontend (Vue 3)
- [x] Initialize Vue project with Vite
- [x] Main layout / page (dark theme, g711.org-inspired)
- [x] File upload component (drag-and-drop, single + ZIP)
- [x] Conversion options panel (format, normalization, bandpass toggles)
- [x] Progress / status display
- [x] Download results panel
- [x] Easter egg mode frontend (session check + Comic Sans rainbow)

## Docker
- [x] Multi-stage Dockerfile (Node → Go → Alpine + ffmpeg)
- [x] docker-compose.yml with env var support (`PORT`, `PROXY_HEADER`, `MAX_UPLOAD_MB`, `CONVERSIONS_DIR`, `EASTER_EGG_IPS`)

## Testing & Verification
- [x] Go backend compiles (`go build`, `go vet`)
- [x] Vue frontend builds (`npm run build`)
- [x] Frontend UI verified in browser
- [ ] End-to-end conversion test with actual audio file
- [ ] Docker build + run smoke test
