<template>
  <div class="app">
    <!-- Background decoration -->
    <div class="bg-decoration">
      <div class="bg-orb orb-1"></div>
      <div class="bg-orb orb-2"></div>
      <div class="bg-orb orb-3"></div>
    </div>

    <!-- Header -->
    <header class="app-header">
      <div class="container">
        <div class="header-content">
          <div class="header-brand">
            <div class="logo">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="url(#logo-gradient)" stroke-width="2">
                <defs>
                  <linearGradient id="logo-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stop-color="#3b82f6" />
                    <stop offset="100%" stop-color="#06b6d4" />
                  </linearGradient>
                </defs>
                <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
                <path d="M15.54 8.46a5 5 0 0 1 0 7.07" />
                <path d="M19.07 4.93a10 10 0 0 1 0 14.14" />
              </svg>
            </div>
            <div>
              <h1>Audio Converter</h1>
              <p class="tagline">Telephony-grade audio conversion powered by ffmpeg</p>
            </div>
          </div>
          <div class="header-badges">
            <span class="badge badge-blue">ffmpeg</span>
            <span class="badge badge-green">8kHz WAV</span>
            <span class="badge badge-purple">G.711</span>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="app-main">
      <div class="container">
        <div class="converter-layout">
          <!-- Left: Upload + Results -->
          <div class="converter-primary">
            <div class="card">
              <FileUpload
                :files="files"
                :disabled="isProcessing"
                @add="addFiles"
                @remove="removeFile"
                @clear="clearFiles"
                @download="downloadFile"
                @set-file-preset="setFilePreset"
              />
            </div>

            <!-- Progress -->
            <ProgressTracker
              :status="overallStatus"
              :progress="overallProgress"
              :doneCount="doneFiles.length"
              :totalCount="files.length"
              :errorCount="errorFiles.length"
            />

            <!-- Results -->
            <ResultsPanel
              :hasDoneFiles="hasDoneFiles"
              :doneCount="doneFiles.length"
              :totalCount="files.length"
              @downloadAll="downloadAllAsZip"
              @downloadFirst="downloadFirstDone"
              @reset="clearFiles"
            />

            <!-- Convert Button -->
            <div v-if="hasFiles && (hasPendingFiles || errorFiles.length > 0)" class="convert-action">
              <button
                class="btn btn-primary btn-lg convert-btn"
                :disabled="isProcessing"
                @click="convertAll"
              >
                <svg v-if="!isProcessing" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polygon points="5 3 19 12 5 21 5 3" />
                </svg>
                <div v-else class="spinner"></div>
                {{ convertButtonLabel }}
              </button>
            </div>
          </div>

          <!-- Right: Options -->
          <div class="converter-sidebar">
            <div class="card">
              <ConversionOptions
                :options="options"
                :formats="formats"
                :prefixes="prefixes"
                :selectedPreset="selectedPreset"
                :presetOptions="presetOptions"
                :disabled="isProcessing"
                @update:options="Object.assign(options, $event)"
                @set-preset="setPreset"
              />
            </div>
          </div>
        </div>

        <!-- Footer info -->
        <footer class="app-footer">
          <div class="footer-info">
            <span>Max file size: 50MB per file</span>
            <span>•</span>
            <span>Supports: MP3, WAV, OGG, FLAC, M4A, AAC</span>
          </div>
        </footer>
      </div>
    </main>

    <!-- Easter egg celebration overlay -->
    <CelebrationOverlay
      v-if="showCelebration"
      :visible="showCelebration"
      @dismiss="dismissCelebration"
    />
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import FileUpload from './components/FileUpload.vue'
import ConversionOptions from './components/ConversionOptions.vue'
import ProgressTracker from './components/ProgressTracker.vue'
import ResultsPanel from './components/ResultsPanel.vue'
import CelebrationOverlay from './components/CelebrationOverlay.vue'
import { useConverter } from './composables/useConverter.js'

const {
  files,
  options,
  formats,
  prefixes,
  selectedPreset,
  presetOptions,
  isProcessing,
  hasFiles,
  hasPendingFiles,
  hasDoneFiles,
  pendingFiles,
  doneFiles,
  errorFiles,
  overallStatus,
  overallProgress,
  addFiles,
  removeFile,
  clearFiles,
  convertAll,
  downloadFile,
  downloadAllAsZip,
  loadFormats,
  loadPrefixes,
  setPreset,
  setFilePreset,
  showCelebration,
  dismissCelebration,
} = useConverter()

const convertButtonLabel = computed(() => {
  if (isProcessing.value) return 'Converting...'
  const pending = pendingFiles.value.length
  const errors = errorFiles.value.length
  const retryCount = pending + errors
  if (doneFiles.value.length > 0 && retryCount > 0) {
    return `Convert ${retryCount} New File${retryCount > 1 ? 's' : ''}`
  }
  return `Convert ${retryCount} File${retryCount > 1 ? 's' : ''}`
})

function downloadFirstDone() {
  const first = doneFiles.value[0]
  if (first) downloadFile(first.id)
}

onMounted(() => {
  loadFormats()
  loadPrefixes()
  checkEasterEgg()
})

async function checkEasterEgg() {
  try {
    const res = await fetch('/api/session')
    if (res.ok) {
      const data = await res.json()
      if (data.easterEgg) {
        document.body.classList.add('easter-egg')
      }
    }
  } catch {
    // Silently ignore
  }
}
</script>

<style scoped>
.app {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
}

/* Background decoration */
.bg-decoration {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}
.bg-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.08;
}
.orb-1 {
  width: 600px;
  height: 600px;
  background: var(--accent-blue);
  top: -200px;
  right: -200px;
}
.orb-2 {
  width: 400px;
  height: 400px;
  background: var(--accent-purple);
  bottom: -100px;
  left: -100px;
}
.orb-3 {
  width: 300px;
  height: 300px;
  background: var(--accent-cyan);
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 24px;
  position: relative;
  z-index: 1;
}

/* Header */
.app-header {
  padding: 32px 0 24px;
  border-bottom: 1px solid var(--border-color);
  background: rgba(11, 14, 23, 0.8);
  backdrop-filter: blur(20px);
  position: sticky;
  top: 0;
  z-index: 10;
}
.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.header-brand {
  display: flex;
  align-items: center;
  gap: 16px;
}
.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: var(--radius-lg);
}
.header-brand h1 {
  font-size: 22px;
  font-weight: 700;
  background: var(--gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.tagline {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 2px;
}
.header-badges {
  display: flex;
  gap: 8px;
}

/* Main */
.app-main {
  padding: 32px 0 64px;
}
.converter-layout {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 24px;
  align-items: start;
}
.converter-primary {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.converter-sidebar .card {
  position: sticky;
  top: 120px;
}
.convert-action {
  display: flex;
  justify-content: center;
}
.convert-btn {
  min-width: 220px;
  justify-content: center;
}

/* Footer */
.app-footer {
  margin-top: 48px;
  padding-top: 24px;
  border-top: 1px solid var(--border-color);
}
.footer-info {
  display: flex;
  gap: 8px;
  justify-content: center;
  font-size: 12px;
  color: var(--text-muted);
}

/* Responsive */
@media (max-width: 768px) {
  .converter-layout {
    grid-template-columns: 1fr;
  }
  .header-badges {
    display: none;
  }
  .converter-sidebar .card {
    position: static;
  }
}
</style>
