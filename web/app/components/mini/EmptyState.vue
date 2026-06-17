<template>
  <div class="empty-state">
    <span class="empty-state__icon">
      <CIcon :name="icon" />
    </span>
    <h2>{{ title }}</h2>
    <p>{{ text }}</p>
    <button v-if="actionLabel" type="button" @click="$emit('action')">
      {{ actionLabel }}
    </button>
  </div>
</template>

<script setup>
defineProps({
  icon: {
    type: String,
    default: 'check',
  },
  title: {
    type: String,
    required: true,
  },
  text: {
    type: String,
    required: true,
  },
  actionLabel: {
    type: String,
    default: '',
  },
});

defineEmits(['action']);
</script>

<style scoped lang="scss">
.empty-state {
  display: grid;
  gap: 18px;
  justify-items: center;
  min-height: 470px;
  padding: 168px 24px 70px;
  color: $annonce-color-ink;
  text-align: center;
  animation: mini-rise-in 240ms $ease-out both;

  &__icon {
    display: grid;
    place-items: center;
    width: 90px;
    height: 90px;
    color: $annonce-color-blue;
    background-color: #ffc261;
    border: 1.5px solid #eba73b;
    border-radius: 26px;
    box-shadow: 0 16px 36px rgba(#0b304b, 0.1);
    animation: mini-soft-pop 230ms $ease-out both;

    :deep(svg) {
      width: 44px;
      height: 44px;
    }
  }

  h2 {
    max-width: 320px;
    font-size: 30px;
    font-weight: 800;
    line-height: 1.12;
  }

  p {
    max-width: 286px;
    font-size: 15px;
    line-height: 1.45;
    color: $annonce-color-ink-muted;
  }

  button {
    min-width: 194px;
    height: 46px;
    padding: 0 18px;
    font-size: 14px;
    font-weight: 800;
    color: $annonce-color-text;
    background-color: $annonce-color-blue;
    border-radius: 23px;
    transition:
      background-color $time-fast $ease-out,
      box-shadow $time-fast $ease-out,
      transform $time-fast $ease-out;

    &:active {
      transform: scale(0.97);
    }
  }
}

@media (hover: hover) {
  .empty-state button:hover {
    background-color: $annonce-color-navy;
    box-shadow: 0 12px 26px rgba(#0b304b, 0.14);
    transform: translate3d(0, -1px, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .empty-state,
  .empty-state__icon {
    animation: none;
  }

  .empty-state button {
    transition-duration: 1ms;
  }
}
</style>
