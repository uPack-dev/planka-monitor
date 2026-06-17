<template>
  <div
    class="s-slider"
    :class="{
      [`s-slider--theme--${theme}`]: theme,
      [`s-slider--overflow--${overflow}`]: overflow,
    }"
  >
    <Swiper
      :key="isBreakpointSmAndUp + isBreakpointMdAndUp"
      class="s-slider__swiper"
      :loop="false"
      :space-between="0"
      :allow-touch-move="true"
      :breakpoints="breakpoints"
      @slide-change="updateEdges"
      @init="onInit"
    >
      <SwiperSlide
        v-for="item in items"
        :key="item.id"
        :lazy="lazy"
        class="s-slider__slide"
      >
        <slot :data="item" />
      </SwiperSlide>
    </Swiper>
  </div>
</template>

<script setup>
import { SLIDER_OVERFLOW, SLIDER_THEME } from '@/configs/sSliderOptions';

defineProps({
  items: {
    type: Array,
    default: () => [],
  },
  breakpoints: {
    type: Object,
    default: () => ({}),
  },
  theme: {
    type: String,
    default: SLIDER_THEME.PRIMARY,
    validator(value) {
      return [...Object.values(SLIDER_THEME)].includes(value);
    },
  },
  lazy: {
    type: Boolean,
    default: false,
  },
  overflow: {
    type: String,
    default: SLIDER_OVERFLOW.VISIBLE,
    validator(value) {
      return [...Object.values(SLIDER_OVERFLOW)].includes(value);
    },
  },
});

const isBeginning = ref(true);
const isEnd = ref(false);
const swiperRef = ref(null);

const updateEdges = (swiper) => {
  isBeginning.value = swiper.isBeginning;
  isEnd.value = swiper.isEnd;
};

const onInit = (swiper) => {
  swiperRef.value = swiper;
};

const prevSlide = () => {
  swiperRef.value.slidePrev?.();
};

const nextSlide = () => {
  swiperRef?.value?.slideNext?.();
};

const { sm: isBreakpointSmAndUp, md: isBreakpointMdAndUp } =
  useCustomBreakpoints();

defineExpose({ swiperRef, prevSlide, nextSlide, isBeginning, isEnd });
</script>

<style lang="scss" scoped>
.s-slider {
  $parent: &;

  display: flex;
  flex-direction: column;
  gap: 56px;
  align-items: center;

  @include media-breakpoint-down(sm) {
    gap: 32px;
  }

  &__swiper {
    width: 100%;
  }

  &__slide {
    width: var(--slider-slide-width, fit-content);
    height: unset;
    margin-right: var(--slider-space-between, 0);
    overflow: hidden;

    &:last-child {
      margin-right: 0;
    }
  }

  &--overflow {
    &--visible {
      #{$parent} {
        &__swiper {
          overflow: visible;
        }
      }
    }

    &--secondary {
      #{$parent} {
        &__swiper {
          overflow: hidden;

          @include media-breakpoint-down(md) {
            overflow: visible;
          }
        }
      }
    }
  }

  // &--theme {
  //   &--secondary {
  //     gap: 72px;

  //     #{$parent} {
  //       &__scrollbar {
  //         background-color: rgba($stroke-color-addition, 0.2);

  //         &:deep() {
  //           .swiper-scrollbar-drag {
  //             background-color: $stroke-color-addition;
  //           }
  //         }
  //       }
  //     }
  //   }

  //   &--tertiary {
  //     gap: 72px;

  //     #{$parent} {
  //       &__arrow {
  //         &--next {
  //           order: 2;
  //         }

  //         &--prev {
  //           order: 1;
  //         }
  //       }
  //     }
  //   }
  // }
}
</style>
