<template>
  <div class="ui-dropdown">
    <div class="ui-dropdown__header" @click="handle">
      <div v-if="title" class="ui-dropdown__title">
        <p class="h3-m-s">{{ title }}</p>
      </div>

      <CIcon
        name="arrow-right"
        class="ui-dropdown__icon"
        :class="{ 'ui-dropdown__icon--active': isActive }"
      />
    </div>

    <ClientOnly>
      <Collapse v-if="$slots.default" :when="isActive">
        <div class="ui-dropdown__description">
          <slot />
        </div>
      </Collapse>
    </ClientOnly>
  </div>
</template>

<script setup>
import { Collapse } from 'vue-collapsed';

const props = defineProps({
  title: {
    type: String,
    default: '',
  },
  initial: {
    type: Boolean,
    default: false,
  },
});

const isActive = ref(props.initial);

const handle = () => {
  isActive.value = !isActive.value;
};
</script>

<style lang="scss" scoped>
.ui-dropdown {
  &__icon {
    width: 14px;
    height: 14px;
    transition: transform $time-fast $ease;

    &--active {
      transform: rotate(90deg);
    }
  }

  &__header {
    display: flex;
    gap: 10px;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    user-select: none;
    transition: color $time-normal $ease-out;

    @include hover {
      color: $actions-color-red;
    }

    @include active {
      color: inherit;
    }
  }

  &__description {
    padding-top: 4px;
  }
}
</style>
