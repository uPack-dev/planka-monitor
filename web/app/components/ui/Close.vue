<script setup>
defineProps({
  size: {
    type: String,
    default: 'sm',
    validator: (size) => ['sm'].includes(size),
  },
});

const emit = defineEmits(['close']);
</script>

<template>
  <button
    aria-label="Close modal"
    class="close-button"
    :class="{ [`close-button--size--${size}`]: !!size }"
    @click="emit('close')"
  >
    <CIcon name="close" class="close-button__icon" />
  </button>
</template>

<style lang="scss" scoped>
.close-button {
  $parent: &;

  color: $actions-color-blue;
  border-radius: 50%;
  transition: color $time-normal ease;

  @include hover {
    color: $actions-color-red;
  }

  @include active {
    color: $actions-color-blue;
  }

  &__icon {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  &--size {
    &--sm {
      width: 22px;
      height: 22px;
    }
  }
}
</style>
