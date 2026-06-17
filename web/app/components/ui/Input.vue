<template>
  <div class="ui-input__wrapper">
    <VeeField
      v-slot="{ errors, field }"
      v-model="value"
      :name="name"
      :rules="rules"
      :validate-on-blur="false"
      :validate-on-change="false"
      :validate-on-input="false"
      :validate-on-model-update="false"
    >
      <label
        class="ui-input"
        :class="{
          [`ui-input--theme--${theme}`]: theme,
          'ui-input--focus': isFocused,
          'ui-input--disabled': disabled,
          'ui-input--error': errors?.length,
          'ui-input--label': label,
          'ui-input--active': value,
        }"
      >
        <input
          v-if="!mask"
          v-bind="field"
          v-model="value"
          class="ui-input__field i1-m-s"
          :type="type"
          :name="name"
          :inputmode="inputMode"
          autocomplete="off"
          :required="isRequired"
          :disabled="disabled"
          :readonly="readonly"
          :aria-required="isRequired"
          :aria-invalid="!!errors?.length"
          :aria-describedby="errors?.length ? `${name}-error` : null"
          :aria-labelledby="label ? `${name}-label` : null"
          :aria-label="label ? null : (correctPlaceholder ?? null)"
          @focus="isFocused = true"
          @blur="isFocused = false"
        />

        <input
          v-else
          v-bind="field"
          v-model="value"
          v-maska="mask"
          class="ui-input__field i1-m-s"
          :type="type"
          :name="name"
          :inputmode="inputMode"
          autocomplete="off"
          :required="isRequired"
          :disabled="disabled"
          :readonly="readonly"
          :aria-required="isRequired"
          :aria-invalid="!!errors?.length"
          :aria-describedby="errors?.length ? `${name}-error` : null"
          :aria-labelledby="correctLabel ? `${name}-label` : null"
          :aria-label="correctLabel ? null : (correctPlaceholder ?? null)"
          @focus="isFocused = true"
          @blur="isFocused = false"
        />

        <span v-if="correctLabel" :id="`${name}-label`" class="ui-input__label">
          <span class="i1-m-s">{{ correctLabel }}</span>
        </span>

        <slot />
      </label>

      <div
        v-if="errors?.length"
        :id="`${name}-error`"
        class="ui-input__msg"
        role="alert"
        aria-live="polite"
      >
        <span class="i2-r-s">
          {{ errors[0] }}
        </span>
      </div>
    </VeeField>
  </div>
</template>

<script setup>
/**
 * @component
 * @example
 * <UiInput
 *   v-model="email"
 *   type="email"
 *   label="Email"
 *   rules="required|email"
 *   placeholder="Enter your email"
 * />
 */

import { vMaska } from 'maska/vue';

import {
  INPUT_THEME,
  INPUT_TYPES,
  INPUT_MODES,
} from '@/configs/uiInputOptions';

const props = defineProps({
  theme: {
    type: String,
    default: INPUT_THEME.PRIMARY,
    validator(value) {
      return [...Object.values(INPUT_THEME)].includes(value);
    },
  },
  type: {
    default: 'text',
    validator(value) {
      return INPUT_TYPES.includes(value);
    },
  },
  inputMode: {
    default: 'text',
    validator(value) {
      return INPUT_MODES.includes(value);
    },
  },
  name: {
    type: String,
    default: '',
  },
  label: {
    type: String,
    default: '',
  },
  modelValue: {
    type: [String, Number],
    default: '',
  },
  rules: {
    type: [String, Object],
    default: '',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  placeholder: {
    type: String,
    default: '',
  },
  readonly: {
    type: Boolean,
    default: false,
  },
  mask: {
    type: Object,
    default: null,
  },
});

const emits = defineEmits(['update:modelValue']);

const isFocused = shallowRef(false);

const value = computed({
  get: () => props.modelValue,
  set: (val) => emits('update:modelValue', val),
});

const isRequired = computed(() => {
  return typeof props.rules === 'string'
    ? props.rules.includes('required')
    : false;
});

const correctLabel = computed(() => {
  if (isRequired.value && props.label) {
    return props.label + '*';
  }

  return props.label;
});

const correctPlaceholder = computed(() => {
  if (isFocused.value) return '';

  if (isRequired.value && props.placeholder) {
    return props.placeholder + '*';
  }

  return props.placeholder;
});
</script>

<style lang="scss" scoped>
.ui-input {
  $parent: &;

  position: relative;
  display: flex;
  width: 100%;
  padding: 12px 25px;
  cursor: text;
  border-radius: 100vw;
  transition: border-color $time-normal ease-in-out;

  &__wrapper {
    width: 100%;
  }

  &__label {
    position: absolute;
    top: 50%;
    left: 22px;
    padding-inline: 3px;
    line-height: 1;
    pointer-events: none;
    user-select: none;
    background-color: transparent;
    border-radius: 20%;
    transform-origin: left center;
    translate: 0 -50%;
    transition: all $time-normal ease-in-out;
  }

  &__field {
    width: 100%;
    color: $background-color-tertiary;
    outline: none;
    background-color: transparent;
    border: none;
    transition: transform $time-normal ease-in-out;

    &::placeholder,
    &:placeholder-shown {
      color: $background-color-tertiary;
    }
  }

  &__msg {
    margin-top: 8px;
    color: $actions-color-red;

    span {
      width: 100%;
    }
  }

  &--label {
    &#{$parent}--focus {
      #{$parent} {
        &__label {
          top: -2px;
          background-color: $color-white;
          scale: 0.8;
        }
      }
    }
  }

  &--active {
    #{$parent} {
      &__label {
        top: -2px;
        background-color: $color-white;
        scale: 0.8;
      }
    }
  }

  &--theme {
    &--primary {
      border: 2px solid rgba($background-color-tertiary, 0.2);

      @include hover {
        border-color: $background-color-tertiary;
      }

      &#{$parent}--focus {
        border-color: $background-color-tertiary;
      }

      &#{$parent}--error {
        border-color: $actions-color-red;
      }

      #{$parent} {
        &__label {
          color: $background-color-tertiary;
        }

        &__field {
          color: $background-color-tertiary;

          &::placeholder,
          &:placeholder-shown {
            color: $background-color-tertiary;
          }
        }
      }
    }

    &--annonce {
      background-color: rgba($color-white, 0.1);
      border: 1.5px solid rgba($color-white, 0.22);
      border-radius: 12px;

      @include hover {
        border-color: rgba($annonce-color-yellow, 0.65);
      }

      &#{$parent}--focus {
        border-color: $annonce-color-yellow;
      }

      &#{$parent}--error {
        border-color: $annonce-color-error;
      }

      #{$parent} {
        &__label {
          color: $annonce-color-muted;
          background-color: $annonce-color-navy;
        }

        &__field {
          color: $annonce-color-text;

          &::placeholder,
          &:placeholder-shown {
            color: rgba($annonce-color-muted, 0.78);
          }
        }
      }
    }

    //   &--secondary {
    //     border: 2px solid rgba($stroke-color-addition, 0.2);

    //     @include hover {
    //       border-color: $stroke-color-addition;
    //     }

    //     &#{$parent}--focus {
    //       border-color: $stroke-color-addition;
    //     }

    //     #{$parent} {
    //       &__label {
    //         color: $text-color-tertiary;
    //       }

    //       &__field {
    //         color: $text-color-invert;

    //         &::placeholder,
    //         &:placeholder-shown {
    //           color: $text-color-tertiary;
    //         }
    //       }
    //     }
    //   }
  }
}
</style>
