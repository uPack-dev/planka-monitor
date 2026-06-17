<script setup>
import socials from '@/configs/data/socials';
</script>

<template>
  <div class="ui-socials">
    <CLinkTag
      v-for="({ icon, link }, index) in [...socials]"
      :key="index"
      :aria-label="`Visit ${icon}`"
      :link="link"
      class="ui-socials__link"
    >
      <CIcon
        class="ui-socials__icon"
        :class="{ [`ui-socials__icon--${icon}`]: !!icon }"
        :name="`socials-${icon}`"
      />
    </CLinkTag>
  </div>
</template>

<style lang="scss" scoped>
.ui-socials {
  display: flex;
  gap: 5px;
  align-items: center;
  justify-content: center;

  &__icon {
    width: 30px;
    height: 30px;

    &--whatsapp {
      width: 26.5px;
      height: 26.5px;
      margin-left: 3px;
    }
  }

  &__link {
    flex-shrink: 0;
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
