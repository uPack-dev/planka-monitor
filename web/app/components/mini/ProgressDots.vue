<template>
  <div class="progress-dots" :aria-label="`Крок ${active} з ${total}`">
    <span
      v-for="index in total"
      :key="index"
      class="progress-dots__dot"
      :class="{
        'progress-dots__dot--active': index === active,
        'progress-dots__dot--completed': index < active,
      }"
    />
  </div>
</template>

<script setup>
defineProps({
  active: {
    type: Number,
    required: true,
  },
  total: {
    type: Number,
    default: 3,
  },
});
</script>

<style scoped lang="scss">
.progress-dots {
  display: flex;
  gap: 6px;
  align-items: center;

  &__dot {
    width: 8px;
    height: 4px;
    background-color: rgba($color-white, 0.18);
    border-radius: 2px;

    &--active {
      width: 24px;
      background-color: $annonce-color-yellow;
    }

    &--completed {
      background-color: rgba($color-white, 0.5);
    }
  }
}
</style>
