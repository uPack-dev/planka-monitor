<template>
  <div class="feature-row">
    <span class="feature-row__icon">
      <CIcon :name="icon" />
    </span>
    <span class="feature-row__body">
      <span class="feature-row__title">{{ title }}</span>
      <span class="feature-row__text">{{ text }}</span>
    </span>
  </div>
</template>

<script setup>
defineProps({
  icon: {
    type: String,
    required: true,
  },
  title: {
    type: String,
    required: true,
  },
  text: {
    type: String,
    required: true,
  },
});
</script>

<style scoped lang="scss">
.feature-row {
  display: grid;
  grid-template-columns: 36px 1fr;
  gap: 14px;
  align-items: start;
  animation: mini-rise-in 220ms $ease-out both;
  animation-delay: var(--motion-delay, 0ms);

  &__icon {
    display: grid;
    place-items: center;
    width: 36px;
    height: 36px;
    color: $annonce-color-yellow;
    background-color: rgba($color-white, 0.13);
    border-radius: 11px;
    transition:
      background-color $time-fast $ease-out,
      transform $time-fast $ease-out;

    :deep(svg) {
      width: 18px;
      height: 18px;
    }
  }

  &__body {
    display: grid;
    gap: 2px;
  }

  &__title {
    font-size: 15.5px;
    font-weight: 700;
    line-height: 20px;
    color: $annonce-color-text;
  }

  &__text {
    font-size: 13.5px;
    line-height: 18px;
    color: rgba($annonce-color-muted, 0.94);
  }
}

@media (hover: hover) {
  .feature-row:hover .feature-row__icon {
    background-color: rgba($annonce-color-yellow, 0.2);
    transform: rotate(-3deg) scale(1.02);
  }
}

@media (prefers-reduced-motion: reduce) {
  .feature-row {
    animation: none;
  }

  .feature-row__icon {
    transition-duration: 1ms;
  }
}
</style>
