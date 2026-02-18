<template>
  <div class="results-panel fade-in" v-if="resultUrl">
    <div class="results-header">
      <div class="results-icon">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
          <polyline points="22 4 12 14.01 9 11.01" />
        </svg>
      </div>
      <div>
        <h3>Ready to Download</h3>
        <p class="results-filename">{{ filename }}</p>
      </div>
    </div>

    <div class="results-actions">
      <button class="btn btn-success btn-lg" @click="$emit('download')">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <polyline points="7 10 12 15 17 10" />
          <line x1="12" y1="15" x2="12" y2="3" />
        </svg>
        Download Converted File
      </button>
      <button class="btn btn-secondary" @click="$emit('reset')">
        Convert Another
      </button>
    </div>
  </div>
</template>

<script setup>
defineProps({
  resultUrl: { type: String, default: null },
  filename: { type: String, default: '' },
})

defineEmits(['download', 'reset'])
</script>

<style scoped>
.results-panel {
  padding: 28px;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.08), rgba(6, 182, 212, 0.08));
  border: 1px solid rgba(16, 185, 129, 0.2);
  border-radius: var(--radius-xl);
  text-align: center;
}
.results-header {
  display: flex;
  align-items: center;
  gap: 16px;
  justify-content: center;
  margin-bottom: 24px;
}
.results-icon {
  color: var(--accent-green);
  flex-shrink: 0;
}
.results-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--accent-green);
}
.results-filename {
  font-size: 13px;
  color: var(--text-secondary);
  font-family: 'SF Mono', 'Fira Code', monospace;
  margin-top: 2px;
}
.results-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  flex-wrap: wrap;
}
</style>
