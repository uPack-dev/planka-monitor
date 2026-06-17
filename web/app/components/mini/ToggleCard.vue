<template>
  <button
    class="toggle-card"
    :class="{ 'toggle-card--active': modelValue }"
    type="button"
    @click="$emit('update:modelValue', !modelValue)"
  >
    <span class="toggle-card__icon">
      <CIcon :name="icon" />
    </span>
    <span class="toggle-card__copy">
      <span class="toggle-card__title">{{ title }}</span>
      <span class="toggle-card__text">{{ text }}</span>
    </span>
    <span class="toggle-card__switch" aria-hidden="true">
      <span class="toggle-card__knob" />
    </span>
  </button>
</template>

<script setup>
defineProps({
  modelValue: {
    type: Boolean,
    required: true,
  },
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

defineEmits(['update:modelValue']);
</script>

<style scoped lang="scss">
.toggle-card {
  display: grid;
  grid-template-columns: 36px 1fr 38px;
  gap: 12px;
  align-items: center;
  width: 100%;
  min-height: 72px;
  padding: 12px;
  color: $annonce-color-muted;
  text-align: left;
  background-color: rgba($color-white, 0.08);
  border: 1.5px solid rgba($color-white, 0.55);
  border-radius: 14px;
  transition:
    background-color $time-normal $ease-out,
    border-color $time-normal $ease-out,
    box-shadow $time-normal $ease-out,
    transform $time-normal $ease-out;
  animation: mini-rise-in 220ms $ease-out both;
  animation-delay: var(--motion-delay, 0ms);

  &:active {
    transform: scale(0.985);
  }

  &__icon {
    display: grid;
    place-items: center;
    width: 36px;
    height: 36px;
    color: $annonce-color-text;
    background-color: rgba($color-white, 0.12);
    border-radius: 10px;
    transition:
      background-color $time-normal $ease-out,
      color $time-normal $ease-out,
      transform $time-normal $ease-out;

    :deep(svg) {
      width: 19px;
      height: 19px;
    }
  }

  &__copy {
    display: grid;
    gap: 3px;
    min-width: 0;
  }

  &__title {
    font-size: 18px;
    font-weight: 700;
    line-height: 1.2;
    color: $annonce-color-text;
  }

  &__text {
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
    font-weight: 500;
    line-height: 1.3;
    color: rgba($annonce-color-muted, 0.92);
    white-space: nowrap;
  }

  &__switch {
    width: 38px;
    height: 22px;
    padding: 3px;
    background-color: rgba($color-white, 0.18);
    border-radius: 11px;
    box-shadow: inset 0 0 0 1px rgba($color-white, 0.42);
    transition: $time-normal $ease-out;
    transition-property: background-color, box-shadow;
  }

  &__knob {
    display: block;
    width: 16px;
    height: 16px;
    background-color: $annonce-color-text;
    border-radius: 50%;
    box-shadow: 0 1px 3px rgba($annonce-color-navy, 0.24);
    transition: translate $time-normal $ease-out;
  }

  &--active {
    background-color: rgba($annonce-color-yellow, 0.12);
    border-color: $annonce-color-yellow;
    box-shadow: 0 10px 28px rgba($annonce-color-yellow, 0.1);

    .toggle-card {
      &__icon {
        color: $annonce-color-text;
        background-color: $annonce-color-yellow;
        transform: rotate(-3deg) scale(1.01);
      }

      &__switch {
        background-color: $annonce-color-yellow;
        box-shadow: none;
      }

      &__knob {
        background-color: $annonce-color-blue;
        translate: 16px 0;
      }
    }
  }
}

@media (hover: hover) {
  .toggle-card:hover {
    background-color: rgba($color-white, 0.12);
    transform: translate3d(0, -1px, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .toggle-card {
    transition-duration: 1ms;
    animation: none;
  }

  .toggle-card__icon,
  .toggle-card__switch,
  .toggle-card__knob {
    transition-duration: 1ms;
  }
}
</style>
