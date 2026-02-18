<template>
  <div class="progress-tracker" v-if="status !== ''">
    <div class="progress-header">
      <div class="progress-status">
        <div v-if="status === 'uploading'" class="status-indicator uploading">
          <div class="spinner"></div>
          <span>Uploading...</span>
        </div>
        <div v-else-if="status === 'converting'" class="status-indicator converting">
          <div class="spinner"></div>
          <span>Converting with ffmpeg...</span>
        </div>
        <div v-else-if="status === 'done'" class="status-indicator done">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
            <polyline points="22 4 12 14.01 9 11.01" />
          </svg>
          <span>Conversion Complete!</span>
        </div>
        <div v-else-if="status === 'error'" class="status-indicator error">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10" />
            <line x1="15" y1="9" x2="9" y2="15" />
            <line x1="9" y1="9" x2="15" y2="15" />
          </svg>
          <span>Conversion Failed</span>
        </div>
      </div>
      <span v-if="status === 'uploading'" class="progress-pct">{{ progress }}%</span>
    </div>

    <div v-if="status === 'uploading' || status === 'converting'" class="progress-bar">
      <div
        class="progress-bar-fill"
        :class="{ indeterminate: status === 'converting' }"
        :style="{ width: status === 'converting' ? '100%' : progress + '%' }"
      ></div>
    </div>

    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </div>
  </div>
</template>

<script setup>
defineProps({
  status: { type: String, default: '' },
  progress: { type: Number, default: 0 },
  errorMessage: { type: String, default: '' },
})
</script>

<style scoped>
.progress-tracker {
  padding: 16px 20px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  animation: fadeIn 0.3s ease;
}
.progress-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.status-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 500;
}
.status-indicator.uploading,
.status-indicator.converting {
  color: var(--accent-blue);
}
.status-indicator.done {
  color: var(--accent-green);
}
.status-indicator.error {
  color: var(--accent-red);
}
.progress-pct {
  font-size: 14px;
  font-weight: 600;
  color: var(--accent-blue);
  font-family: 'SF Mono', 'Fira Code', monospace;
}
.progress-bar-fill.indeterminate {
  animation: indeterminate 1.5s ease-in-out infinite;
  background: linear-gradient(90deg, var(--accent-blue), var(--accent-cyan), var(--accent-blue));
  background-size: 200% 100%;
}
.error-message {
  margin-top: 12px;
  padding: 12px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: var(--radius-md);
  color: var(--accent-red);
  font-size: 13px;
  font-family: 'SF Mono', 'Fira Code', monospace;
  word-break: break-word;
}

@keyframes indeterminate {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
