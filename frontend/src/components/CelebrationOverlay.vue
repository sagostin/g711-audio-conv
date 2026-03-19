<template>
  <Teleport to="body">
    <div v-if="visible" class="celebration-overlay" @click="dismiss">
      <!-- Confetti particles -->
      <div
        v-for="p in particles"
        :key="p.id"
        class="confetti-particle"
        :style="p.style"
      ></div>

      <!-- Spinning star -->
      <img
        src="../assets/your-did-it.png"
        alt="Your did it!"
        class="celebration-star"
      />
    </div>
  </Teleport>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'

const props = defineProps({
  visible: { type: Boolean, default: false },
})

const emit = defineEmits(['dismiss'])

// Generate confetti particles with random properties
const particles = computed(() => {
  if (!props.visible) return []
  const colors = [
    '#ff0000', '#ff8800', '#ffee00', '#33dd00',
    '#0099ff', '#6633ff', '#ff00cc', '#ff69b4',
    '#00ffcc', '#ffaa00', '#44ff44', '#ff4488',
  ]
  const result = []
  for (let i = 0; i < 60; i++) {
    const color = colors[i % colors.length]
    const angle = (i / 60) * 360
    const radians = (angle * Math.PI) / 180
    // Spread distance from center
    const dist = 200 + Math.random() * 300
    const tx = Math.cos(radians) * dist
    const ty = Math.sin(radians) * dist - 200 // bias upward burst, then gravity pulls down
    const size = 6 + Math.random() * 10
    const delay = Math.random() * 0.4
    const duration = 1.8 + Math.random() * 1.2
    const rotation = Math.random() * 720 - 360
    const isCircle = Math.random() > 0.5

    result.push({
      id: i,
      style: {
        '--tx': `${tx}px`,
        '--ty': `${ty}px`,
        '--fall': `${ty + 600}px`,
        '--rot': `${rotation}deg`,
        width: `${size}px`,
        height: `${isCircle ? size : size * 0.5}px`,
        backgroundColor: color,
        borderRadius: isCircle ? '50%' : '2px',
        animationDelay: `${delay}s`,
        animationDuration: `${duration}s`,
      },
    })
  }
  return result
})

let timer = null

onMounted(() => {
  timer = setTimeout(() => {
    emit('dismiss')
  }, 4000)
})

onUnmounted(() => {
  if (timer) clearTimeout(timer)
})

function dismiss() {
  emit('dismiss')
}
</script>

<style scoped>
.celebration-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.7);
  cursor: pointer;
  animation: overlay-fade-in 0.2s ease forwards;
}

@keyframes overlay-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* Spinning star */
.celebration-star {
  width: 280px;
  height: auto;
  animation: star-entrance 2.4s cubic-bezier(0.22, 1, 0.36, 1) forwards;
  filter: drop-shadow(0 0 40px rgba(255, 200, 0, 0.6));
  pointer-events: none;
  z-index: 2;
}

@keyframes star-entrance {
  0% {
    transform: scale(0) rotate(0deg);
    opacity: 0;
  }
  15% {
    opacity: 1;
  }
  100% {
    transform: scale(1) rotate(1080deg);
    opacity: 1;
  }
}

/* Confetti particles */
.confetti-particle {
  position: fixed;
  top: 50%;
  left: 50%;
  z-index: 1;
  pointer-events: none;
  opacity: 0;
  animation: confetti-burst ease-out forwards;
}

@keyframes confetti-burst {
  0% {
    transform: translate(-50%, -50%) translate(0, 0) rotate(0deg);
    opacity: 1;
  }
  30% {
    transform: translate(-50%, -50%) translate(var(--tx), var(--ty)) rotate(var(--rot));
    opacity: 1;
  }
  100% {
    transform: translate(-50%, -50%) translate(var(--tx), var(--fall)) rotate(var(--rot));
    opacity: 0;
  }
}
</style>
