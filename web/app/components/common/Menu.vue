<template>
  <ul class="menu" @touchmove.stop>
    <li v-for="(item, index) in items" :key="index" class="menu__item">
      <CLinkTag
        :link="item.link"
        class="i1-m-s menu__font"
        :aria-label="item.title"
        @click="emit('click')"
      >
        {{ item.title }}
      </CLinkTag>
    </li>
  </ul>
</template>

<script setup>
defineProps({
  items: {
    type: Array,
    default: () => [],
  },
});

const emit = defineEmits(['click']);
</script>

<style lang="scss" scoped>
li.menu__item {
  position: relative;
  user-select: none;
  transition: color 0.3s ease-out;

  &::after {
    position: absolute;
    bottom: 0;
    left: 50%;
    width: 0;
    height: 2px;
    pointer-events: none;
    content: '';
    background: currentcolor;
    transform: translateX(-50%);
    transition: width 0.3s ease-in-out;
    will-change: width;
  }

  &:active::after {
    width: 100%;
  }

  @media (hover: hover) {
    &:hover {
      color: $actions-color-red;

      &::after {
        width: 100%;
      }
    }

    &:active {
      color: inherit;
      transition: color 0.1s ease-out;
    }
  }
}

.menu {
  display: flex;
  flex-shrink: 0;
  gap: 50px;

  @include media-breakpoint-down(md) {
    @include hide-scroll;

    flex-direction: column;
    gap: 20px;
    align-items: center;
    padding-block: 16px;
    margin-block: auto;
    text-transform: capitalize;

    &__item {
      flex-shrink: 0;
      text-align: center;
      letter-spacing: 2%;
    }
  }

  &__font {
    display: block;
    min-height: 100%;
    line-height: 110%;
    text-transform: capitalize;

    @include media-breakpoint-down(md) {
      @include typography-mobile(h2);
    }
  }

  &__item {
    flex-shrink: 0;
  }
}
</style>
