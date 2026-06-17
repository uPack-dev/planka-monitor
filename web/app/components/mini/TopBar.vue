<template>
  <header class="top-bar" :class="{ 'top-bar--compact': !tabs?.length }">
    <div class="top-bar__inner">
      <div class="top-bar__copy">
        <span>{{ eyebrow }}</span>
        <strong>{{ title }}</strong>
      </div>
      <div ref="menuRoot" class="top-bar__actions">
        <button
          v-if="showSearch"
          type="button"
          aria-label="Пошук"
          @click="$emit('search')"
        >
          <CIcon name="search-mini" />
        </button>
        <button
          type="button"
          aria-label="Меню"
          :aria-expanded="menuOpen"
          aria-haspopup="menu"
          @click="toggleMenu"
        >
          <CIcon name="more-mini" />
        </button>
        <Transition name="top-bar-menu">
          <div v-if="menuOpen" class="top-bar__menu" role="menu">
            <button
              type="button"
              role="menuitemcheckbox"
              :aria-checked="motionDisabled"
              @click="toggleMotionPreference"
            >
              <CIcon :name="motionDisabled ? 'check' : 'settings'" />
              <span>{{ motionToggleLabel }}</span>
            </button>
            <button type="button" role="menuitem" @click="restartOnboarding">
              <CIcon name="settings" />
              <span>Перезапуск налаштування</span>
            </button>
          </div>
        </Transition>
      </div>
      <div v-if="tabs?.length" class="top-bar__tabs">
        <UiTabs
          :key="tabsKey"
          :model-value="activeTab"
          :tabs="tabs"
          label-by="label"
          track-by="key"
          variant="annonce-mini"
          @update:model-value="updateActiveTab"
        />
      </div>
    </div>
  </header>
</template>

<script setup>
const props = defineProps({
  eyebrow: {
    type: String,
    default: 'ANNONCE · ADMIN',
  },
  title: {
    type: String,
    required: true,
  },
  modelValue: {
    type: String,
    default: '',
  },
  tabs: {
    type: Array,
    default: () => [],
  },
  showSearch: {
    type: Boolean,
    default: true,
  },
});

const emit = defineEmits(['update:modelValue', 'search', 'restart-onboarding']);

const menuOpen = shallowRef(false);
const menuRoot = shallowRef(null);
const { motionDisabled, toggleMotionDisabled } = useMotionPreference();

const activeTab = computed(
  () => props.tabs.find((tab) => tab.key === props.modelValue) || props.tabs[0],
);
const tabsKey = computed(() => props.tabs.map((tab) => tab.key).join(':'));
const motionToggleLabel = computed(() =>
  motionDisabled.value ? 'Увімкнути анімації' : 'Вимкнути анімації',
);

function updateActiveTab(tab) {
  if (!tab?.key) return;
  emit('update:modelValue', tab.key);
}

function toggleMenu() {
  menuOpen.value = !menuOpen.value;
}

function closeMenu() {
  menuOpen.value = false;
}

function restartOnboarding() {
  closeMenu();
  emit('restart-onboarding');
}

function toggleMotionPreference() {
  toggleMotionDisabled();
  closeMenu();
}

onMounted(() => {
  onClickOutside(menuRoot, closeMenu);
});
</script>

<style scoped lang="scss">
.top-bar {
  position: sticky;
  top: 0;
  z-index: 10;
  min-height: calc(var(--tg-safe-top, 0px) + 128px);
  padding: calc(var(--tg-safe-top, 0px) + 20px) 16px 11px;
  color: $annonce-color-text;
  background-color: $annonce-color-navy;

  &__inner {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 7px 12px;
    width: 100%;

    @include media-breakpoint-up(sm) {
      max-width: 1040px;
      margin: 0 auto;
    }
  }

  &__copy {
    display: grid;
    gap: 7px;
    min-width: 0;

    span {
      font-size: 12px;
      font-weight: 700;
      color: rgba($annonce-color-muted-strong, 0.68);
    }

    strong {
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 20px;
      line-height: 1.2;
      white-space: nowrap;
    }
  }

  &__actions {
    position: relative;
    display: flex;
    gap: 6px;
    align-items: start;
    min-width: fit-content;
    margin-top: 10px;

    button {
      display: grid;
      place-items: center;
      width: 44px;
      height: 44px;
      background-color: rgba($color-white, 0.14);
      border: 1.5px solid rgba($color-white, 0.28);
      border-radius: 14px;
      transition:
        background-color $time-normal $ease-out,
        border-color $time-normal $ease-out,
        box-shadow $time-normal $ease-out,
        transform $time-normal $ease-out;

      &:active {
        transform: scale(0.98);
      }
    }

    :deep(svg) {
      width: 20px;
      height: 20px;
      transition: transform $time-normal $ease-out;
    }
  }

  &__menu {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    z-index: 20;
    width: min(258px, calc(100vw - 32px));
    padding: 6px;
    color: $annonce-color-ink;
    background-color: $color-white;
    border: 1px solid rgba($annonce-color-blue, 0.16);
    border-radius: 8px;
    box-shadow: 0 18px 44px rgb(2 27 43 / 24%);

    button {
      display: flex;
      gap: 10px;
      align-items: center;
      width: 100%;
      height: auto;
      min-height: 44px;
      padding: 10px 12px;
      font-size: 14px;
      font-weight: 700;
      line-height: 1.2;
      color: inherit;
      text-align: left;
      background-color: transparent;
      border: 0;
      border-radius: 6px;
      transition:
        background-color $time-fast $ease-out,
        transform $time-fast $ease-out;

      &:active {
        background-color: rgba($annonce-color-blue, 0.08);
        transform: scale(0.985);
      }
    }

    :deep(svg) {
      flex: 0 0 auto;
      width: 18px;
      height: 18px;
      color: $annonce-color-blue;
    }
  }

  &__tabs {
    grid-column: 1 / -1;
    width: min-content;
    max-width: 100%;
    overflow-x: auto;
    scrollbar-width: none;

    &::-webkit-scrollbar {
      display: none;
    }
  }

  &--compact {
    min-height: calc(var(--tg-safe-top, 0px) + 104px);
  }
}

.top-bar-menu-enter-active,
.top-bar-menu-leave-active {
  transform-origin: top right;
  transition:
    opacity $time-fast $ease,
    filter $time-fast $ease,
    transform $time-fast $ease;
}

.top-bar-menu-enter-from,
.top-bar-menu-leave-to {
  opacity: 0;
  filter: blur(2px);
  transform: translate3d(0, -3px, 0) scale(0.98);
}

@media (hover: hover) {
  .top-bar__actions button:hover {
    background-color: rgba($color-white, 0.2);
    border-color: rgba($color-white, 0.42);
    box-shadow: 0 10px 24px rgba(#031d31, 0.18);
    transform: translate3d(0, -1px, 0);

    :deep(svg) {
      transform: scale(1.02);
    }
  }
}

@media (prefers-reduced-motion: reduce) {
  .top-bar__actions button,
  .top-bar__actions :deep(svg),
  .top-bar__menu button,
  .top-bar-menu-enter-active,
  .top-bar-menu-leave-active {
    transition-duration: 1ms;
  }

  .top-bar-menu-enter-from,
  .top-bar-menu-leave-to {
    filter: none;
    transform: none;
  }
}
</style>
