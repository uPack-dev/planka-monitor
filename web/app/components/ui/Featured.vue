<script setup>
import featured from '@/configs/data/featured';
</script>

<template>
  <div class="ui-featured">
    <CLinkTag
      v-for="({ icon, link }, index) in featured"
      :key="index"
      :aria-label="`Visit ${icon}`"
      :link="link"
      class="ui-featured__link"
    >
      <CIcon class="ui-featured__icon" :name="`socials-${icon}`" />
    </CLinkTag>
  </div>
</template>

<style lang="scss" scoped>
.ui-featured {
  display: flex;
  gap: 20px;
  align-items: center;
  justify-content: center;

  @include media-breakpoint-down(sm) {
    gap: 12px;
  }

  &__icon {
    width: 40px;
    height: 40px;

    @include media-breakpoint-down(sm) {
      width: 35px;
      height: 35px;
    }
  }

  &__link {
    flex-shrink: 0;
    border-radius: 50%;
    box-shadow: 0 4px 4px 0 rgba($color-black, 0.25);
    transition: scale $time-normal $ease-out;
    will-change: scale;

    @include hover {
      scale: 1.1;
    }

    @include active {
      scale: 0.95;
      transition: scale $time-fast $ease-out;
    }
  }
}
</style>
