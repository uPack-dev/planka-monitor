<script setup>
import { Swiper, SwiperSlide } from 'swiper/vue';

const props = defineProps({
  tabs: {
    type: Array,
    default: () => [],
  },
  labelBy: {
    type: String,
    default: 'name',
  },
  trackBy: {
    type: String,
    default: 'slug',
  },
  variant: {
    type: String,
    default: 'shutter',
    validator: (v) =>
      ['shutter', 'shutter-filters', 'plain', 'annonce-mini'].includes(v),
  },
});

const model = defineModel({
  type: Object,
});

const { sm: isBreakpointSmAndUp, md: isBreakpointMdAndUp } =
  useCustomBreakpoints();

const swiperRef = ref();
const tabRefs = ref({});

const activeTab = ref(model.value);
const initialSlide = computed(() =>
  props.tabs.findIndex(
    (tab) => tab[props.trackBy] === activeTab.value?.[props.trackBy],
  ),
);

function onSwiper(swiper) {
  swiperRef.value = swiper;
}

function selectTab(tab) {
  model.value = tab;
  activeTab.value = tab;
}

watch(
  model,
  (tab) => {
    activeTab.value = tab;
  },
  { immediate: true },
);

const shutterStyle = computed(() => {
  const tabEl = tabRefs.value[activeTab.value?.[props.trackBy]];

  if (!tabEl || !swiperRef.value?.el)
    return {
      width: '0px',
      translate: '0px',
    };

  const tabsContainerRef = swiperRef.value.el.children[0];
  const tabRect = tabEl.getBoundingClientRect();
  const containerRect = tabsContainerRef.getBoundingClientRect();

  const tabRelativeLeft = Math.floor(tabRect.left - containerRect.left);
  const tabWidth = Math.floor(tabRect.width);

  return {
    width: `${tabWidth}px`,
    translate: `${tabRelativeLeft}px`,
  };
});

onBeforeUnmount(() => {
  if (swiperRef.value?.__swiper__) {
    swiperRef.value.destroy(true, true);
  }
});
</script>

<template>
  <Swiper
    v-if="tabs?.length > 1"
    :key="isBreakpointSmAndUp + isBreakpointMdAndUp"
    slides-per-view="auto"
    centered-slides
    centered-slides-bounds
    slide-to-clicked-slide
    prevent-interaction-on-transition
    class="tabs-base"
    :class="[`tabs-base--variant--${variant}`]"
    :initial-slide="initialSlide"
    role="tablist"
    @swiper="onSwiper"
  >
    <template #wrapper-end>
      <span
        v-if="['shutter', 'shutter-filters', 'annonce-mini'].includes(variant)"
        :style="shutterStyle"
        class="tabs-base__shutter"
      ></span>
    </template>

    <SwiperSlide
      v-for="tab in tabs"
      :key="tab[trackBy]"
      class="tabs-base__slide"
    >
      <button
        :ref="(el) => (tabRefs[tab[trackBy]] = el)"
        type="button"
        role="tab"
        :class="{
          'tabs-base__tab--active': tab[trackBy] === activeTab?.[trackBy],
        }"
        :tabindex="tab[trackBy] === activeTab?.[trackBy] ? '0' : '-1'"
        class="tabs-base__tab"
        @click="selectTab(tab)"
      >
        <span class="tabs-base__font">
          {{ tab[labelBy] }}
        </span>
      </button>
    </SwiperSlide>
  </Swiper>
</template>

<style lang="scss" scoped>
.tabs-base {
  $parent: &;

  max-width: 100%;
  overflow: visible;
  white-space: nowrap;

  &:deep(.swiper-wrapper) {
    position: relative;
    display: inline-flex;
    width: auto;
    overflow: hidden;
  }

  &__slide {
    display: flex;
    flex-shrink: 0;
    width: auto;
    pointer-events: none;
  }

  &__tab {
    z-index: 1;
    flex-grow: 1;
    flex-shrink: 0;
    text-align: center;
    overflow-wrap: anywhere;
    pointer-events: auto;
    transition: $time-normal $ease-out;
    transition-property: color, background-color, transform;

    &:active {
      transform: scale(0.98);
    }
  }

  &__shutter {
    position: absolute;
    top: 4px;
    bottom: 4px;
    left: 0;
    z-index: -1;
    pointer-events: none;
    background-color: $background-color-tertiary;
    border-radius: inherit;
    transition: $time-normal $ease-out;
    transition-property: box-shadow, translate, width;
    will-change: box-shadow, width, translate;
  }

  &--variant {
    &--shutter {
      #{$parent} {
        &__font {
          @include i2('m', 's');
        }

        &__slide {
          margin-right: 4px;

          &:first-of-type {
            margin-left: 4px;
          }
        }

        &__tab {
          padding: 12px 16px;
          color: $text-color-primary;
          text-transform: uppercase;
        }
      }

      &.swiper-initialized {
        #{$parent} {
          &__tab {
            &--active {
              color: $text-color-invert;
            }
          }
        }
      }

      :deep(.swiper-wrapper) {
        border: 1px solid $stroke-color-primary;
        border-radius: 100px;
      }
    }

    &--shutter-filters {
      #{$parent} {
        &__font {
          @include i2('m', 's');
        }

        &__slide {
          margin-right: 4px;

          &:first-of-type {
            margin-left: 4px;
          }
        }

        &__tab {
          padding: 16px 32px;
          color: $text-color-primary;
          text-transform: uppercase;
        }
      }

      &.swiper-initialized {
        #{$parent} {
          &__tab {
            &--active {
              color: $text-color-invert;
            }
          }
        }
      }

      :deep(.swiper-wrapper) {
        background-color: $color-white;
        border: 1px solid $color-white;
        border-radius: 4px;

        @include media-breakpoint-down(sm) {
          border-color: rgba($stroke-color-secondary, 0.2);
        }
      }
    }

    &--plain {
      #{$parent} {
        &__font {
          @include i1('r', 's');
        }

        &__slide {
          &:not(:first-child) {
            margin-left: 8px;
          }
        }

        &__tab {
          padding: 8px 16px;
          color: $text-color-secondary;
          border-radius: 6px;
          box-shadow: inset 0 0 0 1px $stroke-color-secondary;

          &--active {
            color: $text-color-invert;
            background-color: $background-color-primary;
          }
        }
      }
    }

    &--annonce-mini {
      width: max-content;

      #{$parent} {
        &__shutter {
          top: 3px;
          bottom: 3px;
          z-index: 0;
          background-color: $annonce-color-yellow;
          border-radius: 14px;
          box-shadow: 0 6px 16px rgba($annonce-color-yellow, 0.18);
        }

        &__slide {
          margin: 0;
        }

        &__tab {
          min-width: 58px;
          height: 28px;
          padding: 0 10px;
          font-size: 12.5px;
          font-weight: 700;
          line-height: 16px;
          color: #b4c8d2;
          border-radius: 14px;

          &--active {
            color: $annonce-color-navy;
            transform: translate3d(0, -1px, 0);
          }
        }
      }

      :deep(.swiper-wrapper) {
        align-items: center;
        height: 34px;
        padding: 3px;
        background-color: rgba($color-white, 0.16);
        border: 1.5px solid rgba($color-white, 0.26);
        border-radius: 17px;
      }
    }
  }
}

@media (hover: hover) {
  .tabs-base__tab:hover {
    transform: translate3d(0, -1px, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .tabs-base__tab,
  .tabs-base__shutter {
    transition-duration: 1ms;
  }
}
</style>
