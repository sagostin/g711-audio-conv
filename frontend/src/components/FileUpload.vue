<template>
  <div
    class="upload-zone"
    :class="{ 'drag-over': isDragging, 'has-files': files.length > 0 }"
    @dragover.prevent="onDragOver"
    @dragleave.prevent="onDragLeave"
    @drop.prevent="onDrop"
  >
    <div v-if="files.length === 0" class="upload-prompt">
      <div class="upload-icon">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <polyline points="17 8 12 3 7 8" />
          <line x1="12" y1="3" x2="12" y2="15" />
        </svg>
      </div>
      <h3>Drop audio files here</h3>
      <p>or click to browse — each file is converted independently</p>
      <div class="supported-formats">
        <span class="badge badge-blue">.mp3</span>
        <span class="badge badge-blue">.wav</span>
        <span class="badge badge-blue">.ogg</span>
        <span class="badge badge-blue">.flac</span>
        <span class="badge badge-blue">.m4a</span>
        <span class="badge badge-blue">.aac</span>
      </div>
      <input
        ref="fileInput"
        type="file"
        multiple
        :accept="acceptTypes"
        class="file-input-hidden"
        @change="onFileSelect"
      />
      <button class="btn btn-secondary" @click="$refs.fileInput.click()">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <polyline points="17 8 12 3 7 8" />
          <line x1="12" y1="3" x2="12" y2="15" />
        </svg>
        Browse Files
      </button>
    </div>

    <div v-else class="file-list">
      <div class="file-list-header">
        <h3>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
            <polyline points="14 2 14 8 20 8" />
          </svg>
          {{ files.length }} file{{ files.length > 1 ? 's' : '' }}
        </h3>
        <div class="file-actions">
          <button class="btn btn-secondary" @click="$refs.fileInput.click()" :disabled="disabled">
            + Add More
          </button>
          <button class="btn btn-secondary" @click="$emit('clear')" :disabled="disabled">
            Clear All
          </button>
        </div>
      </div>

      <div class="files-grid">
        <div v-for="f in files" :key="f.id" class="file-item fade-in" :class="'file-status-' + f.status">
          <!-- Status icon -->
          <div class="file-status-icon">
            <svg v-if="f.status === 'pending'" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="icon-pending">
              <circle cx="12" cy="12" r="10" />
              <polyline points="12 6 12 12 16 14" />
            </svg>
            <div v-else-if="f.status === 'converting'" class="spinner-sm"></div>
            <svg v-else-if="f.status === 'done'" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="icon-done">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
              <polyline points="22 4 12 14.01 9 11.01" />
            </svg>
            <svg v-else-if="f.status === 'error'" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="icon-error">
              <circle cx="12" cy="12" r="10" />
              <line x1="15" y1="9" x2="9" y2="15" />
              <line x1="9" y1="9" x2="15" y2="15" />
            </svg>
          </div>

          <div class="file-info">
            <span class="file-name">{{ f.name }}</span>
            <div class="file-meta">
              <span class="file-size">{{ formatSize(f.size) }}</span>
              <span v-if="f.prefix && f.prefix.prefix" class="badge" :class="prefixBadgeClass(f.prefix)">
                {{ f.prefix.description }} ({{ f.prefix.targetDb }}dB)
              </span>
              <span v-else class="badge badge-blue">Standard (-6dB)</span>
              <span v-if="f.status === 'error'" class="file-error-text">{{ f.error }}</span>
            </div>
            <!-- Per-file progress bar during conversion -->
            <div v-if="f.status === 'converting'" class="file-progress">
              <div class="file-progress-bar">
                <div class="file-progress-fill" :style="{ width: f.progress + '%' }"></div>
              </div>
              <span class="file-progress-pct">{{ f.progress }}%</span>
            </div>
            <!-- Audio stats after conversion -->
            <div v-if="f.status === 'done' && f.audioStats" class="audio-stats">
              <div class="stats-row">
                <span class="stats-label">IN</span>
                <span class="stats-tag stats-loudness">{{ f.audioStats.inputLoudness?.toFixed(1) }} LUFS</span>
                <span class="stats-tag stats-peak">Peak {{ f.audioStats.inputPeak?.toFixed(1) }} dB</span>
                <span class="stats-tag stats-lra">LRA {{ f.audioStats.inputLRA?.toFixed(1) }}</span>
              </div>
              <div class="stats-row">
                <span class="stats-label">OUT</span>
                <span class="stats-tag stats-loudness">{{ f.audioStats.outputLoudness?.toFixed(1) }} LUFS</span>
                <span class="stats-tag stats-peak">Peak {{ f.audioStats.outputPeak?.toFixed(1) }} dB</span>
                <span class="stats-tag stats-lra">LRA {{ f.audioStats.outputLRA?.toFixed(1) }}</span>
              </div>
            </div>
          </div>

          <!-- Per-file download button when done -->
          <button v-if="f.status === 'done'" class="file-download" @click="$emit('download', f.id)" title="Download">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
          </button>

          <button class="file-remove" @click="$emit('remove', f.id)" :disabled="disabled" title="Remove file">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
      </div>

      <input
        ref="fileInput"
        type="file"
        multiple
        :accept="acceptTypes"
        class="file-input-hidden"
        @change="onFileSelect"
      />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  files: { type: Array, default: () => [] },
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits(['add', 'remove', 'clear', 'download'])

const isDragging = ref(false)
const fileInput = ref(null)

const acceptTypes = '.mp3,.wav,.ogg,.flac,.m4a,.wma,.aac'

function onDragOver() {
  isDragging.value = true
}

function onDragLeave() {
  isDragging.value = false
}

function onDrop(e) {
  isDragging.value = false
  if (e.dataTransfer.files.length) {
    emit('add', e.dataTransfer.files)
  }
}

function onFileSelect(e) {
  if (e.target.files.length) {
    emit('add', e.target.files)
    e.target.value = ''
  }
}

function formatSize(bytes) {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function prefixBadgeClass(prefix) {
  switch (prefix.label) {
    case 'auto_attendant': return 'badge-green'
    case 'mailbox_greeting': return 'badge-amber'
    case 'hold_music': return 'badge-purple'
    default: return 'badge-blue'
  }
}
</script>

<style scoped>
.upload-zone {
  border: 2px dashed var(--border-color);
  border-radius: var(--radius-xl);
  padding: 40px;
  text-align: center;
  transition: var(--transition);
  cursor: pointer;
  background: rgba(26, 32, 52, 0.5);
}
.upload-zone:hover,
.upload-zone.drag-over {
  border-color: var(--accent-blue);
  background: rgba(59, 130, 246, 0.05);
  box-shadow: var(--shadow-glow);
}
.upload-zone.has-files {
  cursor: default;
  padding: 24px;
  text-align: left;
}
.upload-prompt {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
.upload-icon {
  color: var(--text-muted);
  margin-bottom: 8px;
}
.upload-prompt h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}
.upload-prompt p {
  color: var(--text-muted);
  font-size: 14px;
}
.supported-formats {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: center;
  margin: 8px 0 16px;
}
.file-input-hidden {
  display: none;
}
.file-list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.file-list-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
}
.file-actions {
  display: flex;
  gap: 8px;
}
.file-actions .btn {
  padding: 6px 14px;
  font-size: 13px;
}
.files-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}
.file-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  transition: var(--transition);
}
.file-item:hover {
  border-color: var(--accent-blue);
}
.file-item.file-status-done {
  border-color: rgba(16, 185, 129, 0.3);
  background: rgba(16, 185, 129, 0.03);
}
.file-item.file-status-error {
  border-color: rgba(239, 68, 68, 0.3);
  background: rgba(239, 68, 68, 0.03);
}
.file-item.file-status-converting {
  border-color: rgba(59, 130, 246, 0.3);
}

/* Status icons */
.file-status-icon {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.icon-pending { color: var(--text-muted); }
.icon-done { color: var(--accent-green); }
.icon-error { color: var(--accent-red); }
.spinner-sm {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-blue);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.file-info {
  flex: 1;
  min-width: 0;
}
.file-name {
  display: block;
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.file-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 2px;
  flex-wrap: wrap;
}
.file-size {
  font-size: 12px;
  color: var(--text-muted);
}
.file-error-text {
  font-size: 11px;
  color: var(--accent-red);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 200px;
}

/* Per-file progress */
.file-progress {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
}
.file-progress-bar {
  flex: 1;
  height: 4px;
  background: var(--border-color);
  border-radius: 2px;
  overflow: hidden;
}
.file-progress-fill {
  height: 100%;
  background: var(--gradient-primary);
  border-radius: 2px;
  transition: width 0.3s ease;
}
.file-progress-pct {
  font-size: 11px;
  color: var(--accent-blue);
  font-family: 'SF Mono', 'Fira Code', monospace;
  min-width: 32px;
  text-align: right;
}

/* Action buttons */
.file-download {
  background: none;
  border: none;
  color: var(--accent-green);
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  transition: var(--transition);
  flex-shrink: 0;
}
.file-download:hover {
  background: rgba(16, 185, 129, 0.1);
}
.file-remove {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  transition: var(--transition);
  flex-shrink: 0;
}
.file-remove:hover {
  color: var(--accent-red);
  background: rgba(239, 68, 68, 0.1);
}

/* Audio stats */
.audio-stats {
  margin-top: 6px;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.stats-row {
  display: flex;
  align-items: center;
  gap: 6px;
}
.stats-label {
  font-size: 9px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-muted);
  width: 24px;
  flex-shrink: 0;
  font-family: 'SF Mono', 'Fira Code', monospace;
}
.stats-tag {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 8px;
  font-family: 'SF Mono', 'Fira Code', monospace;
  white-space: nowrap;
}
.stats-loudness {
  background: rgba(59, 130, 246, 0.12);
  color: rgba(96, 165, 250, 0.9);
}
.stats-peak {
  background: rgba(245, 158, 11, 0.12);
  color: rgba(245, 158, 11, 0.9);
}
.stats-lra {
  background: rgba(139, 92, 246, 0.12);
  color: rgba(167, 139, 250, 0.9);
}
</style>
