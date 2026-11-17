<script setup lang="ts">
import Conference from '@/VideoConference.vue'
</script>

<template>
  <div class="app">
    <header class="top-bar">top</header>
    <main class="conference">
      <Conference />
    </main>
    <footer class="bottom-bar">bottom</footer>
    <div class="controls-overlay">controls</div>
  </div>
</template>

<style scoped>
:global(body) {
  margin: 0;
}

.app {
  height: 100dvh; /* This keeps all child elements contained in veiwport without overflowing */
  display: flex;
  flex-direction: column;
  overflow: hidden; /* safety clip for subpixel / future children */
  background: #2a2a2a;
  position: relative; /* anchors .controls-overlay; currently missing */
}

.top-bar {
  flex: 0 0 40px;
  background: #1c1c1c;
}
.bottom-bar {
  flex: 0 0 48px;
  background: #1c1c1c;
}

.conference {
  flex: 1 1 auto; /* takes leftover height; basis auto fine */
  overflow: hidden; /* clip: replaces all the min-* guards */
}

.controls-overlay {
  position: absolute;
  inset: 0;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.2s ease;
}
</style>
