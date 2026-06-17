<script setup>
import { LABEL_THEME } from '@/configs/uiLabelOptions';

defineProps({
  text: {
    type: String,
    default: '',
  },
  theme: {
    type: String,
    default: LABEL_THEME.PRIMARY,
    validator(value) {
      return [...Object.values(LABEL_THEME)].includes(value);
    },
  },
});
</script>

<template>
  <div
    v-if="text"
    class="ui-label"
    :class="{ [`ui-label--theme--${theme}`]: !!theme }"
  >
    <p class="i2-m-g">{{ $tp(text) }}</p>
  </div>
</template>

<style lang="scss" scoped>
.ui-label {
  display: inline-flex;
  align-items: center;
  min-height: 40px;
  padding: 7px 30px;
  text-transform: capitalize;
  border-radius: 100vw;

  @include media-breakpoint-down(md) {
    min-height: 31px;
  }

  &--theme {
    &--primary {
      color: $text-color-primary;
      background-color: $background-color-invert;
    }
  }
}
</style>
