<template>
  <header
    class="header"
    :class="{
      [`header--submenu`]: menuOpen,
      ['header--transparent']: isTransparent,
      [`header--theme--${theme}`]: !!theme,
    }"
  >
    <div class="container header__container">
      <CLinkTag
        class="header__burger"
        :class="{
          'header__burger--open': menuOpen,
        }"
        aria-label="Open menu"
        @click="toggleMenu"
      >
        <nav class="header__menu hidden-tablet">
          <CMenu :items="globalStore.menu.header" />
        </nav>

        <span v-for="n in 3" :key="n" class="header__burger--line" />
      </CLinkTag>
    </div>

    <div
      class="header__submenu submenu"
      :class="{
        'submenu--open': menuOpen,
      }"
    >
      <div class="submenu__container">
        <div class="submenu__content">
          <CMenu :items="globalStore.menu.header" @click="toggleMenuMobile" />
        </div>

        <div class="submenu__footer"></div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { useScrollLock } from '@/composables/useScrollLock';
import { HEADER_TYPE } from '@/configs/headerThemeOptions';

defineProps({
  isTransparent: {
    type: Boolean,
    default: false,
  },
  theme: {
    type: String,
    default: HEADER_TYPE.PRIMARY,
    validator: (v) => Object.values(HEADER_TYPE).includes(v),
  },
});

const globalStore = useGlobalStore();

const route = useRoute();

const scrollLock = useScrollLock();

const menuOpen = ref(false);

const stopWatcher = watch(() => route, closeMenu, { deep: true });

const { md: isBreakpointMdAndUp } = useCustomBreakpoints();

const stopSecondWatcher = watch(() => isBreakpointMdAndUp.value, closeMenu);

onBeforeUnmount(() => {
  stopWatcher();
  stopSecondWatcher();
});

function closeMenu() {
  menuOpen.value = false;
  scrollLock.unlock();
}

function toggleMenu() {
  menuOpen.value = !menuOpen.value;

  if (menuOpen.value) {
    scrollLock.lock();
  } else {
    scrollLock.unlock();
  }
}

function toggleMenuMobile() {
  toggleMenu();

  window.scrollTo({
    top: 0,
    behavior: 'smooth',
  });
}
</script>

<style lang="scss" scoped>
.header {
  $parent: &;

  height: #{$header-height}px;

  @include media-breakpoint-down(md) {
    height: #{$header-height-adaptive}px;
  }

  &__container {
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    height: 100%;
    background-color: inherit;
  }

  &__burger {
    position: relative;
    display: none;
    flex-shrink: 0;
    flex-direction: column;
    align-items: flex-end;
    justify-content: center;
    width: 24px;
    height: 48px;
    cursor: pointer;

    @include media-breakpoint-down(md) {
      display: flex;
    }

    &--line {
      position: absolute;
      top: 50%;
      display: block;
      width: 100%;
      height: 2px;
      margin-top: -1px;
      color: inherit;
      background-color: currentcolor;
      border-radius: 1px;
      transform-origin: center center;
      transition:
        transform $time-normal,
        opacity $time-normal;
      will-change: transform;

      &:first-child {
        transform: translate(0, -8px);
      }

      &:last-child {
        transform: translate(0, 8px);
      }

      &:nth-child(2) {
        width: 75%;
      }
    }

    &--open {
      span:first-child {
        transform: rotate(45deg) translate(0, 0);
      }

      span:nth-child(2) {
        opacity: 0;
      }

      span:last-child {
        transform: rotate(-45deg) translate(0, 0);
      }
    }
  }

  &__menu {
    color: inherit;
  }
}

.submenu {
  display: none;
  background: $gradient-header-submenu;

  @include media-breakpoint-down(md) {
    display: block;
    visibility: hidden;
    height: 100dvh;
    padding: 0 16px 100px;
    transform: translate3d(100%, 0, 1px);
    transition:
      transform $time-normal,
      visibility $time-normal;

    &--open {
      visibility: visible;
      transform: translate3d(0, 0, 1px);
    }
  }

  &__container {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow-y: auto;
    overscroll-behavior: contain;
    touch-action: pan-y;
    -webkit-overflow-scrolling: touch;
  }

  &__content {
    display: flex;
    flex-grow: 1;
    flex-shrink: 0;
    flex-direction: column;
  }

  &__footer {
    display: flex;
    flex-direction: column;
    gap: 10px;
    align-items: flex-start;
    margin-top: 10px;

    & > * {
      flex-shrink: 0;
    }
  }
}
</style>
