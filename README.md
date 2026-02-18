# Audio Converter

Telephony-grade audio conversion tool powered by **ffmpeg**. Converts audio files to formats compatible with PBX/telephony systems (BroadWorks, Asterisk, FreeSWITCH, CUCM, etc.).

## Features

- **Format Conversion** — 8kHz WAV (PCM), G.711 µ-law, G.711 A-law, G.722
- **Auto Normalization** — Based on file prefix:
  - `aa_` (Auto Attendant): -6 dB
  - `mbx_` (Mailbox Greeting): -6 dB
  - `moh_` (Hold Music): -20 dB
- **Bandpass Filter** — Telephony-grade (300Hz–3400Hz), adjustable
- **Bulk ZIP Upload** — Upload ZIP, convert all audio files, download converted ZIP with log
- **Conversion Log** — Detailed per-file conversion results included in ZIP output

## Quick Start

### Docker (Recommended)

```bash
# Copy and edit environment config
cp .env.example .env

# Start with Caddy (SSL auto-provisioned)
docker compose up --build
```

**Local dev** → Open [https://localhost](https://localhost) (self-signed cert)

**Production** → Set `SITE_ADDRESS=audio.yourdomain.com` and `TLS_EMAIL=you@email.com` in `.env`, Caddy will auto-provision Let's Encrypt certs.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SITE_ADDRESS` | `localhost` | Domain for Caddy (e.g. `audio.example.com`) |
| `TLS_EMAIL` | `internal` | Email for Let's Encrypt (`internal` = self-signed) |
| `HTTP_PORT` | `80` | Public HTTP port |
| `HTTPS_PORT` | `443` | Public HTTPS port |
| `PORT` | `8080` | Internal backend port |
| `PROXY_HEADER` | `X-Forwarded-For` | Header for client IP (set by Caddy) |
| `MAX_UPLOAD_MB` | `200` | Maximum upload size in MB |
| `CONVERSIONS_DIR` | `./conversions` | Directory for conversion job storage |
| `EASTER_EGG_IPS` | _(empty)_ | Comma-separated IPs for easter egg mode |

### Development

**Backend:**
```bash
cd backend
go run .
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

Frontend dev server runs on `http://localhost:5173` with API proxy to `http://localhost:8080`.

## API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/formats` | GET | List available output formats |
| `/api/prefixes` | GET | List file prefix configurations |
| `/api/convert` | POST | Convert single audio file |
| `/api/convert/bulk` | POST | Convert ZIP of audio files |
| `/api/session` | GET | Session flags (easter egg check) |

### POST /api/convert

Multipart form fields:
- `file` — Audio file
- `format` — Output format (`wav-pcm`, `wav-ulaw`, `wav-alaw`, `g722`)
- `normalize` — Enable normalization (`true`/`false`)
- `bandpass` — Enable bandpass filter (`true`/`false`)
- `bandpass_low` — Highpass cutoff in Hz (default: 300)
- `bandpass_high` — Lowpass cutoff in Hz (default: 3400)

### POST /api/convert/bulk

Same form fields as above, but `file` must be a `.zip` archive containing audio files.

Returns a ZIP with converted files and `conversion_log.txt`.

## Tech Stack

- **Backend**: Go (standard library + ffmpeg CLI)
- **Frontend**: Vue 3 + Vite
- **Audio**: ffmpeg
- **Reverse Proxy / SSL**: Caddy 2 (auto Let's Encrypt)
- **Deployment**: Docker Compose (Alpine + ffmpeg + Caddy)
