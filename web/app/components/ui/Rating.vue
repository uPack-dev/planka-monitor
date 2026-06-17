<template>
  <div class="rating" :class="{ [`rating--size--${size}`]: size }">
    <div v-for="star in 5" :key="star" class="rating__star">
      <svg
        width="16"
        height="17"
        viewBox="0 0 16 17"
        xmlns="http://www.w3.org/2000/svg"
        class="rating__star-bg"
      >
        <path
          d="M8.34113 13.4298L12.2179 15.9888C12.7179 16.3167 13.3333 15.8289 13.1871 15.2291L12.0641 10.6389C12.0337 10.5116 12.0385 10.378 12.078 10.2536C12.1175 10.129 12.19 10.0186 12.2871 9.93513L15.7639 6.92023C16.2177 6.52837 15.987 5.73667 15.3947 5.69668L10.8565 5.3928C10.7326 5.38531 10.6135 5.34048 10.5138 5.26378C10.4142 5.1871 10.3381 5.08187 10.2949 4.96096L8.6027 0.530607C8.5579 0.402577 8.47627 0.291978 8.36895 0.213825C8.26163 0.135671 8.1338 0.09375 8.00269 0.09375C7.87165 0.09375 7.74382 0.135671 7.6365 0.213825C7.52915 0.291978 7.44754 0.402577 7.40274 0.530607L5.7105 4.96096C5.66737 5.08187 5.5913 5.1871 5.49158 5.26378C5.39186 5.34048 5.27282 5.38531 5.14899 5.3928L0.61071 5.69668C0.0184218 5.73667 -0.212338 6.52837 0.241494 6.92023L3.71828 9.93513C3.81545 10.0186 3.88794 10.129 3.92742 10.2536C3.96688 10.378 3.97171 10.5116 3.94134 10.6389L2.90293 14.8933C2.72601 15.613 3.46444 16.1968 4.05673 15.8049L7.66424 13.4298C7.7654 13.3629 7.88281 13.3274 8.00269 13.3274C8.12264 13.3274 8.23997 13.3629 8.34113 13.4298Z"
          stroke="currentColor"
          fill="none"
        />
      </svg>

      <svg
        width="16"
        height="17"
        viewBox="0 0 16 17"
        :style="{ clipPath: getClipPath(star) }"
        xmlns="http://www.w3.org/2000/svg"
        class="rating__star-fill"
      >
        <path
          d="M8.34113 13.4298L12.2179 15.9888C12.7179 16.3167 13.3333 15.8289 13.1871 15.2291L12.0641 10.6389C12.0337 10.5116 12.0385 10.378 12.078 10.2536C12.1175 10.129 12.19 10.0186 12.2871 9.93513L15.7639 6.92023C16.2177 6.52837 15.987 5.73667 15.3947 5.69668L10.8565 5.3928C10.7326 5.38531 10.6135 5.34048 10.5138 5.26378C10.4142 5.1871 10.3381 5.08187 10.2949 4.96096L8.6027 0.530607C8.5579 0.402577 8.47627 0.291978 8.36895 0.213825C8.26163 0.135671 8.1338 0.09375 8.00269 0.09375C7.87165 0.09375 7.74382 0.135671 7.6365 0.213825C7.52915 0.291978 7.44754 0.402577 7.40274 0.530607L5.7105 4.96096C5.66737 5.08187 5.5913 5.1871 5.49158 5.26378C5.39186 5.34048 5.27282 5.38531 5.14899 5.3928L0.61071 5.69668C0.0184218 5.73667 -0.212338 6.52837 0.241494 6.92023L3.71828 9.93513C3.81545 10.0186 3.88794 10.129 3.92742 10.2536C3.96688 10.378 3.97171 10.5116 3.94134 10.6389L2.90293 14.8933C2.72601 15.613 3.46444 16.1968 4.05673 15.8049L7.66424 13.4298C7.7654 13.3629 7.88281 13.3274 8.00269 13.3274C8.12264 13.3274 8.23997 13.3629 8.34113 13.4298Z"
          fill="currentColor"
          stroke="currentColor"
          stroke-width="1"
        />
      </svg>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  value: {
    type: [String, Number],
    default: '0',
    validator: (value) => value >= 0 && value <= 5,
  },
  size: {
    type: String,
    default: 'sm',
    validator: (value) => ['xs', 'sm', 'md', 'lg'].includes(value),
  },
});

const getClipPath = (starNumber) => {
  const starValue = starNumber - 1;
  const rating = props.value;

  if (rating >= starNumber) {
    return 'inset(0 0 0 0)';
  } else if (rating > starValue && rating < starNumber) {
    const percentage = ((rating - starValue) * 100).toFixed(1);
    return `inset(0 ${100 - percentage}% 0 0)`;
  } else {
    return 'inset(0 100% 0 0)';
  }
};
</script>

<style lang="scss" scoped>
.rating {
  $parent: &;

  display: flex;
  gap: 2px;
  align-items: center;

  &__star {
    position: relative;
    flex-shrink: 0;

    svg {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      image-rendering: crisp-edges;
      image-rendering: -webkit-optimize-contrast;
    }

    &-bg {
      color: $stroke-color-primary;
    }

    &-fill {
      color: $icon-color-star;
    }

    &--empty {
      #{$parent} {
        &__star-fill {
          opacity: 0;
        }
      }
    }

    &--full {
      #{$parent} {
        &__star-bg {
          opacity: 0;
        }
      }
    }
  }

  &__value {
    margin-left: 6px;
    font-size: 14px;
    font-weight: 500;
    color: currentcolor;
  }

  &--size {
    &--xs {
      gap: 1px;

      #{$parent} {
        &__star {
          width: 12px;
          height: 12px;
        }

        &__value {
          margin-left: 4px;
          font-size: 12px;
        }
      }
    }

    &--sm {
      gap: 4px;

      #{$parent} {
        &__star {
          width: 16px;
          height: 16px;

          @include media-breakpoint-down(sm) {
            width: 14px;
            height: 14px;
          }
        }
      }
    }

    &--md {
      gap: 4px;

      #{$parent} {
        &__star {
          width: 20px;
          height: 20px;
        }
      }
    }
  }
}
</style>
