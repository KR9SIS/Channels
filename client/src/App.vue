<script setup lang="ts">
const hangUpCall = () => {
  // Close call and send signalling server notification
  // to other peer, requesting it also closes their connection
  console.error('implement')
}
const elem = document.getElementById('hangup-button')
if (elem != null) {
  elem.addEventListener('click', hangUpCall)
}
console.log(elem)
</script>

<template>
  <div class="flexChild" id="camera-container">
    <div class="camera-box">
      <video id="received_video" autoplay></video>
      <video id="local_video" autoplay muted></video>
      <button id="hangup-button" disabled>Hang Up</button>
    </div>
  </div>
</template>

<style scoped></style>
