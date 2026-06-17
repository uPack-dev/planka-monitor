<template>
  <Teleport to="body">
    <div
      v-if="activeItem"
      class="task-action-sheet"
      :class="{
        'task-action-sheet--open': isOpen,
        'task-action-sheet--dragging': isDragging,
      }"
      @click.self="handleBackdropClick"
    >
      <section
        ref="panelEl"
        class="task-action-sheet__panel"
        role="dialog"
        aria-modal="true"
        aria-label="Дії із задачею"
        :style="{ transform: `translate3d(0, ${panelOffset}px, 0)` }"
      >
        <div
          class="task-action-sheet__drag-zone"
          @pointerdown="startDrag"
          @pointermove="moveDrag"
          @pointerup="finishDrag"
          @pointercancel="cancelDrag"
        >
          <span class="task-action-sheet__handle" aria-hidden="true" />
        </div>
        <header class="task-action-sheet__header">
          <div>
            <h2>{{ activeItem.title }}</h2>
            <p>{{ activeItem.boardName }}</p>
          </div>
          <button
            class="task-action-sheet__notify"
            :class="{
              'task-action-sheet__notify--muted': activeItem.isMuted,
              'task-action-sheet__notify--pending': activeItem.isMutePending,
            }"
            type="button"
            :disabled="activeItem.isMutePending"
            :aria-busy="activeItem.isMutePending ? 'true' : 'false'"
            :aria-label="
              activeItem.isMuted
                ? 'Увімкнути сповіщення задачі'
                : 'Заглушити задачу'
            "
            :aria-pressed="activeItem.isMuted"
            @click.stop="emit('mute', activeItem)"
          >
            <CIcon name="bell-filled-mini" />
          </button>
        </header>

        <div class="task-action-sheet__rows">
          <button
            class="task-action-sheet__row task-action-sheet__row--primary"
            type="button"
            @click="$emit('complete', activeItem)"
          >
            <CIcon name="check-filled-mini" />
            <span>Позначити виконаною</span>
          </button>

          <a
            v-if="workspaceUrl"
            class="task-action-sheet__row"
            :href="workspaceUrl"
            target="_blank"
            rel="noreferrer"
            @click="requestClose"
          >
            <CIcon name="workspace-mini" />
            <span>Відкрити в Workspace</span>
          </a>
          <button v-else class="task-action-sheet__row" type="button" disabled>
            <CIcon name="workspace-mini" />
            <span>Відкрити в Workspace</span>
          </button>

          <button
            class="task-action-sheet__row"
            type="button"
            @click="$emit('comment', activeItem)"
          >
            <CIcon name="message-filled-mini" />
            <span>Залишити коментар</span>
          </button>

          <button class="task-action-sheet__row" type="button" disabled>
            <CIcon name="clock-mini" />
            <span>Відкласти на ...</span>
          </button>

          <button class="task-action-sheet__row" type="button" disabled>
            <CIcon name="clock-mini" />
            <span>Змінити дедлайн</span>
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

const emit = defineEmits(['close', 'complete', 'comment', 'mute']);

const SHEET_TRANSITION_MS = 260;
const DRAG_CLOSE_THRESHOLD = 88;
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

const workspaceUrl = computed(
  () =>
    activeItem.value?.cardWorkspaceUrl || activeItem.value?.workspaceUrl || '',
);
const panelOffset = computed(() => (isOpen.value ? dragOffset.value : 520));

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
  return item?.id || item?.taskId || item?.cardId || '';
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
  if (state && state.id !== event.pointerId) return;
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
.task-action-sheet {
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
    gap: 14px;
    width: 100%;
    max-height: min(78dvh, 420px);
    padding: 52px 14px 20px;
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
    place-items: center;
    height: 52px;
    touch-action: none;
    cursor: grab;

    &:active {
      cursor: grabbing;
    }
  }

  &__handle {
    width: 40px;
    height: 5px;
    background-color: #d9d9d9;
    border-radius: 999px;
  }

  &__header {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 34px;
    gap: 10px;
    align-items: start;
    padding: 0 6px;

    h2 {
      margin: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      font-family: $annonce-font-brand;
      font-size: 17px;
      font-weight: 400;
      line-height: 21px;
      color: $annonce-color-blue;
      white-space: nowrap;
    }

    p {
      margin: 4px 0 0;
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 12px;
      line-height: 16px;
      color: #6a808b;
      white-space: nowrap;
    }
  }

  &__notify {
    display: grid;
    place-items: center;
    width: 34px;
    height: 34px;
    color: #2e8c5a;
    opacity: 1;

    &:disabled {
      opacity: 0.62;
    }

    :deep(svg) {
      width: 22px;
      height: 21px;
    }
  }

  &__notify--muted {
    color: #8b989f;
  }

  &__notify--pending {
    pointer-events: none;
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
    font-size: 15px;
    font-weight: 800;
    line-height: 20px;
    color: $annonce-color-ink;
    text-align: left;
    background-color: rgba(#ffffff, 0.72);
    border-radius: 12px;
    transition:
      background-color $time-fast $ease-out,
      box-shadow $time-fast $ease-out,
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
      color: #2e8c5a;
    }
  }

  &--open {
    pointer-events: auto;
    background-color: rgba(#0b304b, 0.45);

    .task-action-sheet__row {
      animation: mini-rise-in 170ms $ease-out both;
    }

    .task-action-sheet__row:nth-child(2) {
      animation-delay: 12ms;
    }

    .task-action-sheet__row:nth-child(3) {
      animation-delay: 24ms;
    }

    .task-action-sheet__row:nth-child(4) {
      animation-delay: 36ms;
    }

    .task-action-sheet__row:nth-child(5) {
      animation-delay: 48ms;
    }
  }
}

.task-action-sheet--dragging .task-action-sheet__panel {
  transition: none;
}

@media (hover: hover) {
  .task-action-sheet__row:not(:disabled):hover {
    background-color: $color-white;
    box-shadow: 0 10px 22px rgba(#0b304b, 0.08);
    transform: translate3d(0, -1px, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .task-action-sheet,
  .task-action-sheet__panel,
  .task-action-sheet__row {
    transition-duration: 1ms;
    animation: none;
  }
}
</style>
