<template>
  <div class="ui-dropdown">
    <div class="ui-dropdown__header" @click="emit('click')">
      <div v-if="question" class="ui-dropdown__title">
        <p class="s1-m-s" v-html="question"></p>
      </div>

      <div
        class="ui-dropdown__icon"
        :class="{ ['ui-dropdown__icon--active']: isActive }"
      >
        <span class="ui-dropdown__icon-part"></span>
        <span class="ui-dropdown__icon-part"></span>
      </div>
    </div>

    <ClientOnly>
      <Collapse v-if="answer" :when="isActive">
        <div class="ui-dropdown__description">
          <p
            class="ui-dropdown__font ui-dropdown__font--description s2-r-s"
            v-html="answer"
          />
        </div>
      </Collapse>
    </ClientOnly>
  </div>
</template>

<script setup>
import { Collapse } from 'vue-collapsed';

defineProps({
  answer: {
    type: String,
    default: '',
  },
  question: {
    type: String,
    default: '',
  },
  isActive: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['click']);
</script>

<style lang="scss" scoped>
.ui-dropdown {
  $parent: &;

  &__icon {
    position: relative;
    flex-shrink: 0;
    width: 40px;
    height: 40px;

    @include media-breakpoint-down(md) {
      width: 30px;
      height: 30px;
    }

    &-part {
      position: absolute;
      top: 50%;
      left: 50%;
      width: 100%;
      height: 0;
      border-block: 1px solid $actions-color-red;
      translate: -50% -50%;
      transition: all $time-fast $ease-in-out;

      &:first-child {
        transform: rotate(-90deg);
      }
    }

    &--active {
      #{$parent} {
        &__icon-part {
          &:first-child {
            opacity: 0.5;
            transform: rotate(0deg);
          }
        }
      }
    }
  }

  &__header {
    display: flex;
    gap: 32px;
    align-items: center;
    justify-content: space-between;
    padding-inline: 10px;
    cursor: pointer;
    user-select: none;
  }

  &__title {
    max-width: 685px;
    padding-block: 20px;
    color: $text-color-primary;
  }

  &__description {
    padding-inline: 10px;
    padding-bottom: 20px;
    color: $text-color-secondary;
  }

  &__subtitle,
  &__description {
    max-width: 90%;
  }
}
</style>
