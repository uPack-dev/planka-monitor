<script setup></script>

<template>
  <div class="part-privacy-policy">
    <p class="s3-m-s">Privacy Policy</p>

    <p class="s3-m-s">© 2025</p>
  </div>
</template>

<style lang="scss" scoped>
.part-privacy-policy {
  display: flex;
  flex-direction: column;
  gap: 10px;
  text-align: right;

  @include media-breakpoint-down(sm) {
    gap: 0;
    text-align: center;
  }
}
</style>
