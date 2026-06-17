<template>
  <CLinkTag
    :link="link"
    :target="computedTarget"
    :type="typeButton"
    :aria-label="title"
    :rel="rel"
    class="ui-button"
    :class="{
      [`ui-button--theme ui-button--theme--${theme}`]: theme,
      [`ui-button--size ui-button--size--${size}`]: size,
      'ui-button--disabled': disabled,
      'ui-button--stretched': stretchContent,
      'ui-button--loading': isLoading,
    }"
    @click="onButtonClick"
  >
    <AFade>
      <span v-if="isLoading" class="ui-button__wrapper">
        <span class="ui-button__spinner"></span>
      </span>
    </AFade>

    <template v-if="icon">
      <span
        v-if="
          typeof icon === 'string' &&
          icon_position === BUTTON_ICON_POSITION.LEFT
        "
        class="ui-button__icon ui-button__icon--left"
      >
        <CIcon :name="icon" />
      </span>

      <span
        v-else-if="icon?.inline && icon_position === BUTTON_ICON_POSITION.LEFT"
        class="ui-button__icon ui-button__icon--left"
        v-html="icon.inline"
      />
    </template>

    <span v-if="title" class="i1-m-s ui-button__font">
      {{ title }}
    </span>

    <template v-if="icon">
      <span
        v-if="
          typeof icon === 'string' &&
          icon_position === BUTTON_ICON_POSITION.RIGHT
        "
        class="ui-button__icon ui-button__icon--right"
      >
        <CIcon :name="icon" />
      </span>

      <span
        v-else-if="icon?.inline && icon_position === BUTTON_ICON_POSITION.RIGHT"
        class="ui-button__icon ui-button__icon--right"
        v-html="icon.inline"
      />
    </template>
  </CLinkTag>
</template>

<script setup>
import {
  BUTTON_ACTION,
  BUTTON_ICON_POSITION,
  BUTTON_SIZE,
  BUTTON_THEME,
} from '@/configs/uiButtonOptions';
import { useModal } from 'vue-final-modal';
import CLinkTag from '@/components/common/CLinkTag.vue';

const props = defineProps({
  title: {
    type: String,
    default: '',
  },
  typeButton: {
    type: String,
    default: 'button',
  },
  link: {
    type: String,
    default: '',
  },
  icon: {
    type: [Object, String],
    default: null,
  },
  target: {
    type: String,
    default: '',
  },
  icon_position: {
    type: String,
    default: BUTTON_ICON_POSITION.RIGHT,
    validator(value) {
      return [...Object.values(BUTTON_ICON_POSITION)].includes(value);
    },
  },
  theme: {
    type: String,
    default: BUTTON_THEME.PRIMARY,
    validator(value) {
      return [...Object.values(BUTTON_THEME)].includes(value);
    },
  },
  action: {
    type: String,
    default: BUTTON_ACTION.LINK,
  },
  size: {
    type: String,
    default: BUTTON_SIZE.SM,
    validator(value) {
      return [...Object.values(BUTTON_SIZE)].includes(value);
    },
  },
  stretchContent: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  isLoading: {
    type: Boolean,
    default: false,
  },
  rel: {
    type: String,
    default: '',
  },
});

const emit = defineEmits(['click']);

const modalConfig = [
  {
    name: BUTTON_ACTION.CALLBACK,
    component: resolveComponent('LazyModalsCallback'),
  },
];

const computedTarget = computed(() =>
  props.action === BUTTON_ACTION.LINK_EXTERNAL ? '_blank' : props.target,
);

const computedModalConfig = computed(() => {
  return modalConfig.find((config) => config.name === props.action);
});

const modal = ref(null);

onMounted(() => {
  if (computedModalConfig.value) {
    modal.value = useModal({
      component: computedModalConfig.value.component,
    });
  }
});

function onButtonClick() {
  if (modal.value) {
    modal.value.open();
    return;
  }

  emit('click');
}
</script>

<style scoped lang="scss">
.ui-button {
  $parent: &;

  position: relative;
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 2px solid transparent;
  border-radius: 100vw;
  transition:
    background-color $time-normal $ease-in-out,
    border-color $time-normal $ease-in-out,
    box-shadow $time-normal $ease-in-out,
    color $time-normal $ease-in-out,
    scale $time-normal $ease-in-out;
  will-change: scale, transform;

  &::before {
    position: absolute;
    inset: 0;
    pointer-events: none;
    content: '';
    background: linear-gradient(
      115deg,
      transparent 0%,
      rgba($color-white, 0.28) 44%,
      transparent 72%
    );
    opacity: 0;
    transform: translate3d(-120%, 0, 0);
    transition:
      opacity $time-normal $ease-out,
      transform $time-slow $ease-out;
  }

  &__font {
    position: relative;
    z-index: 1;
    line-height: 28px;

    @include media-breakpoint-down(sm) {
      line-height: 23px;
    }
  }

  &__icon {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    width: 20px;
    height: 20px;
    transition: transform $time-normal $ease-out;
  }

  &--theme {
    &--primary {
      color: $text-color-invert;
      background-color: $actions-color-red;
      border-color: $actions-color-red;

      @include hover {
        color: $actions-color-red;
        background-color: $actions-color-primary;
        scale: 1.05;
      }

      @include active {
        color: $actions-color-blue;
        background-color: $actions-color-primary;
        border-color: $actions-color-blue;
        scale: 0.95;
      }

      @media (hover: none) {
        &:active {
          color: $actions-color-red;
          background-color: $actions-color-primary;
          border-color: $actions-color-red;
          scale: 1;
          transition: $time-fast ease;
        }
      }
    }

    &--secondary {
      color: $actions-color-primary;
      background-color: $actions-color-blue;
      border-color: $actions-color-blue;
    }

    &--outline {
      color: $actions-color-primary;
      background-color: transparent;
      border-color: $actions-color-primary;
    }

    &--secondary,
    &--outline {
      @include hover {
        color: $actions-color-blue;
        background-color: $actions-color-primary;
        border-color: $actions-color-blue;
        scale: 1.05;
      }

      @include active {
        color: $actions-color-blue;
        background-color: $actions-color-primary;
        border-color: $actions-color-blue;
        scale: 0.95;
      }

      @media (hover: none) {
        &:active {
          color: $actions-color-blue;
          background-color: $actions-color-primary;
          border-color: $actions-color-blue;
          scale: 1;
          transition: $time-fast ease;
        }
      }
    }

    &--annonce {
      color: $annonce-color-navy;
      background-color: $annonce-color-yellow-action;
      border-color: $annonce-color-yellow-action;
      border-radius: 14px;
      box-shadow: 0 10px 18px rgb(122 71 10 / 16%);

      @include hover {
        color: $annonce-color-navy-dark;
        background-color: $annonce-color-yellow;
        border-color: $annonce-color-yellow;
        box-shadow: 0 14px 28px rgb(122 71 10 / 20%);
        scale: 1.02;
      }

      @include active {
        scale: 0.98;
      }

      #{$parent} {
        &__font {
          font-size: 14px;
          font-weight: 700;
          line-height: 20px;
        }
      }
    }

    &--annonce-outline {
      color: $annonce-color-muted;
      background-color: rgba($color-white, 0.08);
      border-color: rgba($annonce-color-muted, 0.55);
      border-radius: 14px;

      @include hover {
        color: $annonce-color-text;
        background-color: rgba($color-white, 0.13);
        border-color: rgba($annonce-color-yellow, 0.85);
        scale: 1.01;
      }
    }

    &--annonce-ghost {
      color: $annonce-color-muted;
      background-color: transparent;
      border-color: transparent;
      border-radius: 14px;

      @include hover {
        color: $annonce-color-text;
        background-color: rgba($color-white, 0.08);
      }
    }
  }

  &--disabled {
    color: rgba($text-color-disable, 0.4);
    pointer-events: none;
    background-color: rgba($text-color-disable, 0.3);
    border-color: rgba($text-color-disable, 0.4);
  }

  &--size {
    &--sm {
      padding: 10px 35px;

      @include media-breakpoint-down(md) {
        padding: 5px 20px;
      }
    }

    &--lg {
      min-height: 52px;
      padding: 14px 28px;
    }
  }

  &--stretched {
    justify-content: space-between;
  }

  &--loading {
    pointer-events: none;

    #{$parent} {
      &__wrapper {
        position: absolute;
        top: 50%;
        left: 50%;
        opacity: 1;
        transform: translate(-50%, -50%);
      }

      &__spinner {
        display: block;
        width: 25px;
        height: 25px;
        border: 4px solid transparent;
        border-top-color: $color-white;
        border-radius: 50%;
        opacity: 1;
        animation: spin 0.8s linear infinite;
      }
    }

    span {
      opacity: 0;
    }
  }
}

@media (hover: hover) {
  .ui-button:hover::before {
    opacity: 1;
    transform: translate3d(120%, 0, 0);
  }

  .ui-button:hover .ui-button__icon {
    transform: translate3d(2px, 0, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .ui-button,
  .ui-button::before,
  .ui-button__icon {
    transition-duration: 1ms;
  }

  .ui-button:hover::before {
    opacity: 0;
    transform: none;
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
