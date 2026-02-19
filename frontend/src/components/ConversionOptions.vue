<template>
  <div class="options-panel">
    <h3 class="options-title">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
      </svg>
      Conversion Options
    </h3>

    <!-- Output Format -->
    <div class="form-group">
      <label class="form-label">Output Format</label>
      <select
        class="form-select"
        :value="options.format"
        @change="$emit('update:options', { ...options, format: $event.target.value })"
        :disabled="disabled"
      >
        <option v-for="f in formats" :key="f.id" :value="f.id">
          {{ f.label }}
        </option>
      </select>
    </div>

    <!-- Normalization Toggle -->
    <div class="toggle-wrapper">
      <div class="toggle-label">
        <span class="toggle-label-text">Audio Normalization</span>
        <span class="toggle-label-desc">Auto-adjusts based on file prefix (aa_: -6dB, moh_: -20dB)</span>
      </div>
      <label class="toggle">
        <input
          type="checkbox"
          :checked="options.normalize"
          @change="$emit('update:options', { ...options, normalize: $event.target.checked })"
          :disabled="disabled"
        />
        <span class="toggle-slider"></span>
      </label>
    </div>

    <!-- Bandpass Filter Toggle -->
    <div class="toggle-wrapper">
      <div class="toggle-label">
        <span class="toggle-label-text">Bandpass Filter</span>
        <span class="toggle-label-desc">Apply telephony-grade bandpass (300Hz–3400Hz)</span>
      </div>
      <label class="toggle">
        <input
          type="checkbox"
          :checked="options.bandpass"
          @change="$emit('update:options', { ...options, bandpass: $event.target.checked })"
          :disabled="disabled"
        />
        <span class="toggle-slider"></span>
      </label>
    </div>

    <!-- Bandpass Range (shown when enabled) -->
    <Transition name="slide">
      <div v-if="options.bandpass" class="bandpass-range">
        <div class="range-inputs">
          <div class="form-group">
            <label class="form-label">Low Cut (Hz)</label>
            <input
              type="number"
              class="form-input"
              :value="options.bandpassLow"
              @input="$emit('update:options', { ...options, bandpassLow: parseFloat($event.target.value) || 300 })"
              min="20"
              max="8000"
              step="10"
              :disabled="disabled"
            />
          </div>
          <div class="range-separator">—</div>
          <div class="form-group">
            <label class="form-label">High Cut (Hz)</label>
            <input
              type="number"
              class="form-input"
              :value="options.bandpassHigh"
              @input="$emit('update:options', { ...options, bandpassHigh: parseFloat($event.target.value) || 3400 })"
              min="20"
              max="8000"
              step="10"
              :disabled="disabled"
            />
          </div>
        </div>
        <div class="range-hint">
          Standard telephony: 300Hz – 3400Hz
        </div>
      </div>
    </Transition>

    <!-- Prefix Reference -->
    <div class="prefix-reference">
      <div class="prefix-title">File Prefix Guide</div>
      <div class="prefix-grid">
        <div v-for="p in prefixes" :key="p.prefix" class="prefix-item">
          <span class="badge" :class="prefixBadgeClass(p)">{{ p.prefix }}</span>
          <span>{{ p.description }}</span>
          <span class="prefix-db">{{ p.targetDb }} dB</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  options: { type: Object, required: true },
  formats: { type: Array, default: () => [] },
  prefixes: { type: Array, default: () => [] },
  disabled: { type: Boolean, default: false },
})

defineEmits(['update:options'])

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
.options-panel {
  padding: 4px 0;
}
.options-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 20px;
  color: var(--text-primary);
}
.bandpass-range {
  padding: 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  margin-bottom: 8px;
}
.range-inputs {
  display: flex;
  align-items: flex-end;
  gap: 12px;
}
.range-inputs .form-group {
  flex: 1;
  margin-bottom: 0;
}
.range-separator {
  color: var(--text-muted);
  padding-bottom: 10px;
  font-weight: 300;
}
.range-hint {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 10px;
  text-align: center;
}
.prefix-reference {
  margin-top: 20px;
  padding: 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}
.prefix-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 12px;
}
.prefix-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.prefix-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--text-secondary);
}
.prefix-db {
  margin-left: auto;
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 12px;
  color: var(--accent-cyan);
}

/* Slide transition */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.25s ease;
  overflow: hidden;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  max-height: 0;
  margin-bottom: 0;
  padding: 0 16px;
}
.slide-enter-to,
.slide-leave-from {
  max-height: 200px;
}
</style>
