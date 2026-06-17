<template>
  <div class="ui-textarea__wrapper">
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
        class="ui-textarea"
        :class="{
          [`ui-textarea--theme--${theme}`]: theme,
          'ui-textarea--focus': isFocused,
          'ui-textarea--disabled': disabled,
          'ui-textarea--error': errors?.length,
          'ui-textarea--label': label,
          'ui-textarea--active': value,
          'ui-textarea--no-resize': !resize,
        }"
      >
        <textarea
          v-if="!mask"
          v-bind="field"
          v-model="value"
          class="ui-textarea__field i1-m-s"
          :name="name"
          :rows="rows"
          :style="{ resize: resize ? 'vertical' : 'none' }"
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
        ></textarea>

        <textarea
          v-else
          v-bind="field"
          v-model="value"
          v-maska="mask"
          class="ui-textarea__field i1-m-s"
          :name="name"
          :rows="rows"
          :style="{ resize: resize ? 'vertical' : 'none' }"
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
        ></textarea>

        <span
          v-if="correctLabel"
          :id="`${name}-label`"
          class="ui-textarea__label"
        >
          <span class="i1-m-s">{{ correctLabel }}</span>
        </span>

        <slot />
      </label>

      <div
        v-if="errors?.length"
        :id="`${name}-error`"
        class="ui-textarea__msg"
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
 * @component UiTextarea
 */

import { computed, shallowRef } from 'vue';
import { Field as VeeField } from 'vee-validate';
import { vMaska } from 'maska/vue';

import { INPUT_THEME } from '@/configs/uiInputOptions';

const props = defineProps({
  name: {
    type: String,
    required: true,
  },
  modelValue: {
    type: [String, Number],
    default: '',
  },
  label: {
    type: String,
    default: '',
  },
  placeholder: {
    type: String,
    default: '',
  },
  theme: {
    type: String,
    default: INPUT_THEME.PRIMARY,
    validator(value) {
      return Object.values(INPUT_THEME).includes(value);
    },
  },
  rules: {
    type: [String, Object],
    default: '',
  },
  rows: {
    type: [String, Number],
    default: 2,
  },
  resize: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
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
    return `${props.label}*`;
  }
  return props.label;
});

const correctPlaceholder = computed(() => {
  if (isFocused.value) return '';

  if (isRequired.value && props.placeholder) {
    return `${props.placeholder}*`;
  }

  return props.placeholder;
});
</script>

<style lang="scss" scoped>
.ui-textarea {
  $parent: &;
  $padding-y: 12px;
  $padding-x: 25px;

  position: relative;
  display: flex;
  width: 100%;
  padding: $padding-y $padding-x;
  cursor: text;
  border-radius: 30px;
  transition: border-color $time-normal ease-in-out;

  &__wrapper {
    width: 100%;
  }

  &__label {
    position: absolute;
    top: $padding-y;
    left: 22px;
    max-width: calc(100% - 44px);
    padding-inline: 3px;
    line-height: 1;
    pointer-events: none;
    user-select: none;
    background-color: transparent;
    border-radius: 20%;
    transform-origin: left center;
    transition: all $time-normal ease-in-out;
  }

  &__field {
    @include scrollbar;

    width: 100%;
    line-height: 1;
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
          opacity: 0;
        }
      }
    }
  }

  &--active {
    #{$parent} {
      &__label {
        opacity: 0;
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
  }
}
</style>
