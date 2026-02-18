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
      <h3>Drop audio files or ZIP here</h3>
      <p>or click to browse</p>
      <div class="supported-formats">
        <span class="badge badge-blue">.mp3</span>
        <span class="badge badge-blue">.wav</span>
        <span class="badge badge-blue">.ogg</span>
        <span class="badge badge-blue">.flac</span>
        <span class="badge badge-blue">.m4a</span>
        <span class="badge badge-blue">.aac</span>
        <span class="badge badge-purple">.zip</span>
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
          {{ files.length }} file{{ files.length > 1 ? 's' : '' }} selected
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
        <div v-for="(f, i) in files" :key="i" class="file-item fade-in" :style="{ animationDelay: i * 50 + 'ms' }">
          <div class="file-icon">
            <svg v-if="f.name.endsWith('.zip')" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 8v13H3V3h13" />
              <path d="M16 3v5h5" />
              <path d="M9 7h1M9 9h1M9 11h1M9 13h1M9 15h1M9 17h1" />
            </svg>
            <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 18V5l12-2v13" />
              <circle cx="6" cy="18" r="3" />
              <circle cx="18" cy="16" r="3" />
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
            </div>
          </div>
          <button class="file-remove" @click="$emit('remove', i)" :disabled="disabled" title="Remove file">
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

const emit = defineEmits(['add', 'remove', 'clear'])

const isDragging = ref(false)
const fileInput = ref(null)

const acceptTypes = '.mp3,.wav,.ogg,.flac,.m4a,.wma,.aac,.zip'

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
  max-height: 280px;
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
.file-icon {
  color: var(--accent-cyan);
  flex-shrink: 0;
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
}
.file-size {
  font-size: 12px;
  color: var(--text-muted);
}
.file-remove {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: var(--transition);
  flex-shrink: 0;
}
.file-remove:hover {
  color: var(--accent-red);
  background: rgba(239, 68, 68, 0.1);
}
</style>
