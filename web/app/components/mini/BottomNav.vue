<template>
  <nav class="bottom-nav" aria-label="Mini app navigation">
    <button
      v-for="item in items"
      :key="item.key"
      class="bottom-nav__item"
      :class="{ 'bottom-nav__item--active': active === item.key }"
      type="button"
      @click="$emit('update:active', item.key)"
    >
      <CIcon :name="item.icon" />
      <span>{{ item.label }}</span>
    </button>
  </nav>
</template>

<script setup>
const props = defineProps({
  active: {
    type: String,
    required: true,
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
});

defineEmits(['update:active']);

const baseItems = [
  { key: 'tasks', label: 'Задачі', icon: 'nav-tasks' },
  { key: 'feed', label: 'Стрічка', icon: 'nav-feed' },
  { key: 'profile', label: 'Профіль', icon: 'nav-profile' },
  { key: 'admin', label: 'Адмін', icon: 'nav-admin' },
];

const items = computed(() =>
  props.isAdmin ? baseItems : baseItems.filter((item) => item.key !== 'admin'),
);
</script>

<style scoped lang="scss">
.bottom-nav {
  position: fixed;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 200;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  height: calc(var(--tg-safe-bottom, 0px) + 90px);
  padding: 18px 15px var(--tg-safe-bottom, 0);
  color: #b4c8d2;
  /* stylelint-disable-next-line property-no-vendor-prefix -- Needed for iOS WebView long-press selection. */
  -webkit-user-select: none;
  user-select: none;
  -webkit-touch-callout: none;
  background-color: $annonce-color-navy;

  &::before {
    position: absolute;
    top: 0;
    right: 18px;
    left: 18px;
    height: 1px;
    content: '';
    background: linear-gradient(
      90deg,
      transparent,
      rgba($annonce-color-yellow, 0.46),
      transparent
    );
  }

  &__item {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 76px;
    min-width: 0;
    height: 54px;
    padding: 7px 4px 8px;
    font-size: 12.5px;
    font-weight: 500;
    line-height: 16px;
    border-radius: 16px;
    transition:
      color $time-normal $ease-out,
      background-color $time-normal $ease-out,
      transform $time-normal $ease-out;
    will-change: transform;

    &:active {
      transform: scale(0.98);
    }

    :deep(svg) {
      --stroke-0: #9eb7c4;
      --fill-0: #9eb7c4;

      width: 22px;
      height: 22px;
      transition:
        transform $time-normal $ease-out,
        filter $time-normal $ease-out;
    }

    &--active {
      font-weight: 600;
      color: $annonce-color-text;
      background-color: #265577;
      box-shadow: inset 0 1px 0 rgba($color-white, 0.08);
      animation: mini-soft-pop 180ms $ease-out;

      :deep(svg) {
        --stroke-0: #ffc261;
        --fill-0: #ffc261;

        filter: drop-shadow(0 3px 8px rgba(#ffc261, 0.24));
        transform: translate3d(0, -1px, 0);
      }
    }
  }
}

@media (hover: hover) {
  .bottom-nav__item:hover {
    color: $annonce-color-text;
    background-color: rgba($color-white, 0.08);
    transform: translate3d(0, -1px, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .bottom-nav__item {
    transition-duration: 1ms;
    animation: none;

    :deep(svg) {
      transform: none;
      transition-duration: 1ms;
    }
  }
}
</style>
