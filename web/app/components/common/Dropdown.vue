<script setup>
const props = defineProps({
  opened: {
    type: Boolean,
    default: false,
  },
  position: {
    type: String,
    default: 'bottom-start',
  },
  autoFlip: {
    type: Boolean,
    default: true,
  },
  triggerEvent: {
    type: String,
    default: 'click',
    validator: (value) => ['click', 'hover'].includes(value),
  },
  referenceRef: {
    type: Object,
    default: undefined,
  },
  hoverHideDelay: {
    type: Number,
    default: 300,
  },
  gap: {
    type: Number,
    default: 0,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  escToClose: {
    type: Boolean,
    default: false,
  },
  contentFullWidth: {
    type: Boolean,
    default: false,
  },
});

const visible = ref(props.opened);

watch(
  () => props.opened,
  (opened) => {
    if (props.disabled) return;

    visible.value = opened;
  },
);

const rootRef = ref(null);

function open() {
  if (props.disabled || visible.value) return;

  visible.value = true;
}

function close() {
  if (props.disabled || !visible.value) return;

  visible.value = false;
}

const referenceElement = computed(() => props.referenceRef || rootRef.value);

// <editor-fold desc="Handling events">
const triggerEvents = {
  hover: 'hover',
  click: 'click',
};
let hideTimeout = null;

// click
function onTriggerClick() {
  if (props.triggerEvent !== triggerEvents.click) return;

  if (visible.value) {
    close();
  } else {
    open();
  }
}

onMounted(() => {
  if (props.triggerEvent === triggerEvents.click) {
    onClickOutside(rootRef, close);
  }
});

// hover
function onPointerEnter() {
  if (props.triggerEvent === triggerEvents.hover) {
    clearTimeout(hideTimeout);
    open();
  }
}

function onPointerLeave() {
  if (props.triggerEvent === triggerEvents.hover) {
    hideTimeout = setTimeout(close, props.hoverHideDelay);
  }
}

// escape
function onEsc() {
  if (visible.value && props.escToClose) close();
}
// </editor-fold>

defineExpose({ open, close, visible });
</script>

<template>
  <div
    ref="rootRef"
    class="dropdown"
    @pointerenter="onPointerEnter"
    @pointerleave="onPointerLeave"
    @keydown.esc="onEsc"
  >
    <div class="dropdown__trigger" @click="onTriggerClick">
      <slot name="trigger" :visible="visible" />
    </div>

    <Transition name="fade">
      <CFloating
        v-if="visible"
        class="dropdown__content"
        :class="{ 'dropdown__content--full-width': contentFullWidth }"
        :reference-ref="referenceElement"
        :position="position"
        :auto-flip="autoFlip"
        :gap="gap"
      >
        <slot name="content" :close="close" />
      </CFloating>
    </Transition>
  </div>
</template>

<style lang="scss" scoped>
.dropdown {
  position: relative;

  &__content {
    z-index: 100;

    &--full-width {
      width: 100%;
    }
  }
}
</style>
