<template>
  <MiniAppShell :theme="theme">
    <div class="state state--center">
      <div class="state__mark" :class="{ 'state__mark--muted': muted }">
        {{ mark }}
      </div>
      <h1 class="state__title">{{ title }}</h1>
      <p class="state__text">{{ text }}</p>
      <p v-if="error" class="screen-error">{{ error }}</p>
    </div>
  </MiniAppShell>
</template>

<script setup>
defineProps({
  theme: {
    type: String,
    default: 'dark',
  },
  mark: {
    type: String,
    default: 'A',
  },
  title: {
    type: String,
    required: true,
  },
  text: {
    type: String,
    required: true,
  },
  error: {
    type: String,
    default: '',
  },
  muted: {
    type: Boolean,
    default: false,
  },
});
</script>

<style scoped lang="scss">
.state {
  display: grid;
  gap: 16px;
  justify-items: center;
  text-align: center;

  &__mark {
    display: grid;
    place-items: center;
    width: 72px;
    height: 72px;
    font-family: $annonce-font-brand;
    font-size: 28px;
    color: $annonce-color-blue;
    background-color: $annonce-color-yellow;
    border-radius: 22px;
    animation: mini-soft-pop 230ms $ease-out both;

    &--muted {
      color: $annonce-color-muted;
      background-color: rgba($color-white, 0.12);
    }
  }

  &__title {
    max-width: 320px;
    font-size: 26px;
    font-weight: 700;
    line-height: 1.15;
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 35ms;
  }

  &__text {
    max-width: 320px;
    font-size: 15px;
    line-height: 1.45;
    color: rgba($annonce-color-muted, 0.9);
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 20ms;
  }

  &--center {
    margin-block: auto;
  }
}

.screen-error {
  font-size: 13px;
  line-height: 1.35;
  color: $annonce-color-error;
  text-align: center;
}

@media (prefers-reduced-motion: reduce) {
  .state__mark,
  .state__title,
  .state__text {
    animation: none;
  }
}
</style>
