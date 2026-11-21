<script setup lang="ts">
import { ref, watchEffect } from "vue";

const {
    src,
    nickname,
    muted = false
} = defineProps<{
    src: MediaStream | null;
    nickname: string;
    muted?: boolean;
}>();
const videoRef = ref<HTMLVideoElement | null>(null);

// runs on mount (ref appears) and whenever src changes
watchEffect(() => {
    if (videoRef.value) {
        videoRef.value.srcObject = src ?? null;
    }
});
</script>

<template>
    <video
        v-if="src"
        ref="videoRef"
        class="player"
        playsinline
        autoplay
        :muted="muted"
    />
    <div v-else class="player nickname">{{ nickname }}</div>
</template>

<style scoped>
.player {
    width: 100%;
    aspect-ratio: 4 / 3;
    display: block;
    border: 2px solid;
    border-radius: 8px;
    background: #2a2a2a;
}
.nickname {
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 2rem;
}
</style>
