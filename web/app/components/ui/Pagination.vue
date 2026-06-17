<template>
  <div
    v-if="totalItems > itemsPerPage"
    ref="containerRef"
    class="ui-pagination"
    :class="{ 'ui-pagination--disabled': isDisabled }"
  >
    <component
      :is="component"
      v-model="page"
      :total-items="totalItemsNumber"
      :items-per-page="itemsPerPageNumber"
      :max-pages-shown="maxShowCurrent"
    >
      <template #prev-button>
        <CIcon
          name="chevron-left"
          class="ui-pagination__icon ui-pagination__icon--prev"
        />
      </template>

      <template #next-button>
        <CIcon
          name="chevron-left"
          class="ui-pagination__icon ui-pagination__icon--next"
        />
      </template>
    </component>
  </div>
</template>

<script setup>
import { useWindowSize } from '@vueuse/core';
import { VueAwesomePaginate } from 'vue-awesome-paginate';

const { width } = useWindowSize();

const props = defineProps({
  totalItems: {
    type: [Number, String],
    default: 0,
  },
  itemsPerPage: {
    type: [Number, String],
    default: 5,
  },
  scrollToTop: {
    type: Boolean,
    default: false,
  },
  scrollToId: {
    type: String,
    default: '',
  },
  modelValue: {
    type: Number,
    default: 1,
  },
  isDisabled: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['update:modelValue']);

const isMounted = ref(false);

const itemsPerPageNumber = computed(() => {
  return Number(props.itemsPerPage) || 0;
});

const totalItemsNumber = computed(() => {
  return Number(props.totalItems) || 0;
});

const maxShowCurrent = computed(() => {
  if (!isMounted.value) return 0;
  return width.value < 1024 ? 3 : 5;
});

const page = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});

const component = computed(() =>
  isMounted.value ? VueAwesomePaginate : 'div',
);

const containerRef = useTemplateRef('containerRef');

const scroll = (top) =>
  setTimeout(() => {
    window.scrollTo({ top: top - 100, behavior: 'smooth' });
  }, 0);

const stopWatcher = watch(
  () => page.value,
  () => {
    if (!isMounted.value) return;

    if (props.scrollToTop) {
      scroll(100);
      return;
    }

    if (props.scrollToId) {
      const element = document.getElementById(props.scrollToId);
      if (element) {
        const rect = element.getBoundingClientRect();
        const absoluteTop = rect.top + window.scrollY;
        scroll(absoluteTop);
        return;
      }
    }

    const parent = containerRef.value?.parentElement;
    if (parent) {
      const rect = parent.getBoundingClientRect();
      const absoluteTop = rect.top + window.scrollY;
      scroll(absoluteTop);
    }
  },
);

onBeforeUnmount(stopWatcher);

onMounted(() => (isMounted.value = true));
</script>

<style lang="scss">
.pagination-container {
  display: flex;
  gap: 8px;

  @include media-breakpoint-down(md) {
    gap: 4px;
  }
}
</style>

<style lang="scss">
$button-size: 35;
$button-size-sm: 35;

.paginate-buttons {
  $parent: &;

  display: flex;
  align-items: center;
  justify-content: center;
  width: #{$button-size}px;
  height: #{$button-size}px;
  border-radius: 100px;
  transition:
    background-color $time-normal ease,
    color $time-normal ease;

  @include media-breakpoint-down(md) {
    width: #{$button-size-sm}px;
    height: #{$button-size-sm}px;
  }

  &.active-page {
    color: $text-color-invert;
    background-color: $actions-color-red;
  }

  &.back-button,
  &.next-button,
  &.number-buttons,
  &.last-button,
  &.first-button,
  &.starting-breakpoint-button,
  &.ending-breakpoint-button {
    @include i1('r', 'o');

    line-height: 1;

    &:hover {
      color: $text-color-invert;
      background-color: $actions-color-red;

      svg {
        color: $text-color-invert;
      }
    }
  }
}

.ui-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: #{$button-size}px;
  transition: opacity $time-normal $ease-in-out;

  @include media-breakpoint-down(md) {
    min-height: #{$button-size-sm}px;
  }

  &__icon {
    width: 8px;
    height: 11px;
    color: $icon-color-primary;
    transition: color $time-normal $ease;

    &--next {
      transform: rotate(180deg);
    }
  }

  &--disabled {
    pointer-events: none;
    opacity: 0.5;
  }
}
</style>
