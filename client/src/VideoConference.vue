<script setup lang="ts">
import { onMounted, ref } from 'vue'

import VideoPlayer from '@/VideoPlayer.vue'

const nickname = localStorage.getItem('nickname') ?? 'Caller'

const selfViewStream = ref<MediaStream | null>(null)

async function toggleSelfView() {
  // TODO: Refactor to independently toggle video and voice
  if (!selfViewStream?.value) {
    selfViewStream.value = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    return
  }
  selfViewStream.value.getTracks().forEach((track) => {
    track.stop()
  })
  selfViewStream.value = null
}

onMounted(() => {
  toggleSelfView()
})
</script>

<template>
  <!-- TODO: move button to controls panel -->
  <button @click="toggleSelfView">Toggle Self View</button>
  <div class="stage">
    <div class="self-view" v-if="selfViewStream">
      <VideoPlayer :src="selfViewStream" :nickname="nickname" :muted="true" />
    </div>
  </div>
</template>

<style scoped>
.stage {
  position: relative;
  height: 100%; /* parent <main> is a definite-height flex item */
}

.self-view {
  position: absolute;
  right: 1rem;
  bottom: 1rem;
  z-index: 10;
  width: 240px; /* TODO: make responsive */
}
</style>
