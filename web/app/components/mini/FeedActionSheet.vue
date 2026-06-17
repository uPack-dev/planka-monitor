<template>
  <Teleport to="body">
    <div
      v-if="activeItem"
      class="feed-action-sheet"
      :class="{
        'feed-action-sheet--open': isOpen,
        'feed-action-sheet--dragging': isDragging,
      }"
      @click.self="handleBackdropClick"
    >
      <section
        ref="panelEl"
        class="feed-action-sheet__panel"
        role="dialog"
        aria-modal="true"
        aria-label="Дії зі стрічкою"
        :style="{ transform: `translate3d(0, ${panelOffset}px, 0)` }"
      >
        <div
          class="feed-action-sheet__drag-zone"
          @pointerdown="startDrag"
          @pointermove="moveDrag"
          @pointerup="finishDrag"
          @pointercancel="cancelDrag"
        >
          <span class="feed-action-sheet__handle" aria-hidden="true" />
        </div>

        <div class="feed-action-sheet__rows">
          <button
            class="feed-action-sheet__row"
            :class="{
              'feed-action-sheet__row--primary': !activeItem.isMuted,
              'feed-action-sheet__row--pending': activeItem.isMutePending,
            }"
            type="button"
            :disabled="!activeItem.cardId || activeItem.isMutePending"
            :aria-busy="activeItem.isMutePending ? 'true' : 'false'"
            :aria-pressed="!activeItem.isMuted"
            @click.stop="emit('mute', activeItem)"
          >
            <CIcon name="bell-filled-mini" />
            <span>{{ notificationLabel }}</span>
          </button>

          <a
            v-if="workspaceUrl"
            class="feed-action-sheet__row"
            :href="workspaceUrl"
            target="_blank"
            rel="noreferrer"
            @click="requestClose"
          >
            <CIcon name="workspace-mini" />
            <span>Відкрити проект в Workspace</span>
          </a>
          <button v-else class="feed-action-sheet__row" type="button" disabled>
            <CIcon name="workspace-mini" />
            <span>Відкрити проект в Workspace</span>
          </button>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<script setup>
const props = defineProps({
  item: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(['close', 'mute']);

const SHEET_TRANSITION_MS = 260;
const DRAG_CLOSE_THRESHOLD = 70;
const BACKDROP_CLOSE_GUARD_MS = 450;

const activeItem = shallowRef(null);
const isOpen = shallowRef(false);
const isDragging = shallowRef(false);
const dragOffset = shallowRef(0);
const dragState = shallowRef(null);
const panelEl = shallowRef(null);
let openFrame = null;
let closeTimer = null;
let suppressBackdropCloseUntil = 0;

const notificationLabel = computed(() =>
  activeItem.value?.isMuted
    ? 'Увімкнути сповіщення проекту'
    : 'Вимкнути сповіщення проекту',
);
const workspaceUrl = computed(
  () =>
    activeItem.value?.projectUrl ||
    activeItem.value?.cardWorkspaceUrl ||
    activeItem.value?.workspaceUrl ||
    '',
);
const panelOffset = computed(() => (isOpen.value ? dragOffset.value : 260));

watch(
  () => props.item,
  async (nextItem) => {
    clearTimeout(closeTimer);
    cancelOpenFrame();

    if (nextItem) {
      const isSameItem =
        isOpen.value && sameActionItem(activeItem.value, nextItem);
      activeItem.value = nextItem;
      dragOffset.value = 0;
      isDragging.value = false;
      dragState.value = null;
      if (isSameItem) {
        return;
      }

      isOpen.value = false;
      await nextTick();
      suppressBackdropCloseUntil = Date.now() + BACKDROP_CLOSE_GUARD_MS;
      openWithAnimation();
      return;
    }

    if (activeItem.value) {
      closeWithoutEmit();
    }
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  clearTimeout(closeTimer);
  cancelOpenFrame();
});

function requestClose() {
  closeWithAnimation(true);
}

function handleBackdropClick() {
  if (Date.now() < suppressBackdropCloseUntil) return;

  requestClose();
}

function closeWithoutEmit() {
  closeWithAnimation(false);
}

function closeWithAnimation(shouldEmit) {
  clearTimeout(closeTimer);
  cancelOpenFrame();
  isDragging.value = false;
  dragState.value = null;
  dragOffset.value = 0;
  isOpen.value = false;
  closeTimer = setTimeout(() => {
    activeItem.value = null;
    if (shouldEmit) {
      emit('close');
    }
  }, SHEET_TRANSITION_MS);
}

function openWithAnimation() {
  panelEl.value?.getBoundingClientRect();

  if (typeof requestAnimationFrame !== 'function') {
    isOpen.value = true;
    return;
  }

  openFrame = requestAnimationFrame(() => {
    openFrame = null;
    isOpen.value = true;
  });
}

function cancelOpenFrame() {
  if (openFrame === null || typeof cancelAnimationFrame !== 'function') return;

  cancelAnimationFrame(openFrame);
  openFrame = null;
}

function sameActionItem(left, right) {
  return (
    actionItemKey(left) !== '' && actionItemKey(left) === actionItemKey(right)
  );
}

function actionItemKey(item) {
  return item?.id || item?.cardId || '';
}

function startDrag(event) {
  if (event.button && event.button !== 0) return;
  isDragging.value = true;
  dragState.value = {
    id: event.pointerId,
    startY: event.clientY,
  };
  capturePointer(event);
}

function moveDrag(event) {
  const state = dragState.value;
  if (!state || state.id !== event.pointerId) return;

  const dy = event.clientY - state.startY;
  dragOffset.value = Math.max(0, dy);
}

function finishDrag(event) {
  const state = dragState.value;
  if (!state || state.id !== event.pointerId) return;

  releasePointer(event);
  dragState.value = null;
  isDragging.value = false;

  if (dragOffset.value >= DRAG_CLOSE_THRESHOLD) {
    requestClose();
    return;
  }

  dragOffset.value = 0;
}

function cancelDrag(event) {
  const state = dragState.value;
  if (state && event.pointerId !== state.id) return;
  releasePointer(event);
  dragState.value = null;
  isDragging.value = false;
  dragOffset.value = 0;
}

function capturePointer(event) {
  try {
    event.currentTarget.setPointerCapture?.(event.pointerId);
  } catch {
    // The drag still works if capture is unavailable.
  }
}

function releasePointer(event) {
  try {
    event.currentTarget.releasePointerCapture?.(event.pointerId);
  } catch {
    // Some browsers release capture automatically.
  }
}
</script>

<style scoped lang="scss">
.feed-action-sheet {
  position: fixed;
  inset: 0 0 calc(var(--tg-safe-bottom, 0px) + 90px);
  z-index: 30;
  display: flex;
  align-items: end;
  pointer-events: none;
  background-color: rgba(#0b304b, 0);
  transition: background-color $time-normal $ease-out;

  &__panel {
    position: relative;
    display: grid;
    width: 100%;
    min-height: 150px;
    max-height: min(48dvh, 240px);
    padding: 36px 14px 20px;
    overflow-y: auto;
    color: $annonce-color-ink;
    /* stylelint-disable-next-line property-no-vendor-prefix -- Needed for iOS WebView long-press selection. */
    -webkit-user-select: none;
    user-select: none;
    -webkit-touch-callout: none;
    background-color: $annonce-color-cream;
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -22px 58px -28px rgba(#0b304b, 0.46);
    transition:
      border-radius $time-normal $ease-out,
      transform $time-normal $ease-out;
    will-change: transform;
  }

  &__drag-zone {
    position: absolute;
    top: 0;
    right: 0;
    left: 0;
    display: grid;
    place-items: start center;
    height: 36px;
    padding-top: 12px;
    touch-action: none;
    cursor: grab;

    &:active {
      cursor: grabbing;
    }
  }

  &__handle {
    width: 38px;
    height: 4px;
    background-color: #d9d9d9;
    border-radius: 999px;
  }

  &__rows {
    display: grid;
    gap: 6px;
  }

  &__row {
    display: grid;
    grid-template-columns: 24px minmax(0, 1fr);
    gap: 12px;
    align-items: center;
    min-height: 44px;
    padding: 0 14px;
    overflow: hidden;
    font-family: Ubuntu, Inter, sans-serif;
    font-size: 15px;
    font-weight: 500;
    line-height: 20px;
    color: $annonce-color-ink;
    text-align: left;
    background-color: rgba(#ffffff, 0.72);
    border-radius: 12px;
    transition:
      opacity $time-fast $ease-out,
      transform $time-fast $ease-out;

    &:disabled {
      color: rgba($annonce-color-ink, 0.45);
      opacity: 1;
    }

    span {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    :deep(svg) {
      justify-self: center;
      width: 16px;
      height: 16px;
      color: $annonce-color-blue;
    }
  }

  &__row--primary {
    background-color: #d2e2d4;

    :deep(svg) {
      width: 18px;
      height: 18px;
      color: #2e8c5a;
    }
  }

  &__row--pending {
    pointer-events: none;
    opacity: 0.7;
  }

  &--open {
    pointer-events: auto;
    background-color: rgba(#0b304b, 0.45);

    .feed-action-sheet__row {
      animation: mini-rise-in 170ms $ease-out both;
    }

    .feed-action-sheet__row:nth-child(2) {
      animation-delay: 15ms;
    }
  }
}

.feed-action-sheet--dragging .feed-action-sheet__panel {
  transition: none;
}

@media (hover: hover) {
  .feed-action-sheet__row:not(:disabled):hover {
    background-color: $color-white;
    box-shadow: 0 10px 22px rgba(#0b304b, 0.08);
    transform: translate3d(0, -1px, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .feed-action-sheet,
  .feed-action-sheet__panel,
  .feed-action-sheet__row {
    transition-duration: 1ms;
    animation: none;
  }
}
</style>
