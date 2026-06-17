<template>
  <article
    class="task-card"
    :class="{
      'task-card--dragging': isDragging,
      'task-card--ready': dragOffset >= swipeThreshold,
    }"
    @contextmenu.prevent="handleContextMenu"
  >
    <div class="task-card__complete-cue" aria-hidden="true">
      <CIcon name="check" />
      <span>Виконано</span>
    </div>

    <button
      class="task-card__main"
      type="button"
      :style="{ transform: `translate3d(${dragOffset}px, 0, 0)` }"
      @click="handleClick"
      @pointerdown="startPointer"
      @pointermove="movePointer"
      @pointerup="finishPointer"
      @pointercancel="cancelPointer"
      @pointerleave="cancelMousePointer"
    >
      <span class="task-card__meta">
        <span>{{ timeLabel }}</span>
        <strong>{{ relativeLabel }}</strong>
      </span>
      <span class="task-card__title">{{ item.title }}</span>
      <span class="task-card__bottom">
        <span class="task-card__board">
          <i :style="{ backgroundColor: boardColor }" />
          {{ item.boardName }}
        </span>
        <span v-if="item.members?.length" class="task-card__avatars">
          <MiniAvatar
            v-for="member in visibleMembers"
            :key="member.id || member.username"
            class="task-card__avatar"
            :user="member"
            size="xs"
            bordered
          />
        </span>
      </span>
    </button>
  </article>
</template>

<script setup>
import {
  formatClockTime,
  parseValidDate,
  relativeDeadlineLabel,
} from '@/utils/miniDate';

const props = defineProps({
  item: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(['open', 'complete', 'menu']);
const telegram = useTelegramWebApp();

const LONG_PRESS_MS = 420;
const TAP_SLOP = 8;
const SWIPE_REVEAL_WIDTH = 54;
const CONTEXT_MENU_SUPPRESS_MS = 800;
const HAPTIC_FALLBACK_DELAY_MS = 120;

const dragOffset = shallowRef(0);
const isDragging = shallowRef(false);
const pointerState = shallowRef(null);
const didGesture = shallowRef(false);
let longPressTimer = null;
let hapticFallbackTimer = null;
let suppressContextMenuUntil = 0;

const swipeThreshold = SWIPE_REVEAL_WIDTH;
const visibleMembers = computed(() => props.item.members?.slice(0, 3) || []);
const dueDate = computed(() => parseValidDate(props.item.dueAt));
const timeLabel = computed(() => {
  if (!dueDate.value) return 'Без дати';
  return formatClockTime(dueDate.value);
});
const relativeLabel = computed(() => {
  if (!dueDate.value) return 'Дедлайн не задано';
  return relativeDeadlineLabel(dueDate.value);
});
const boardColor = computed(() => {
  if (props.item.boardColor) return props.item.boardColor;
  const palette = ['#ffc261', '#2e8c5a', '#275f88', '#d94d67', '#5995c1'];
  const seed = [...(props.item.boardName || '')].reduce(
    (sum, char) => sum + char.charCodeAt(0),
    0,
  );
  return palette[seed % palette.length];
});

onBeforeUnmount(() => {
  clearLongPressTimer();
  clearHapticFallbackTimer();
});

function startPointer(event) {
  if (event.button && event.button !== 0) return;

  const shouldSuppressNativeLongPress =
    (event.pointerType === 'touch' || event.pointerType === 'pen') &&
    event.cancelable;
  if (shouldSuppressNativeLongPress) {
    event.preventDefault();
  }

  clearLongPressTimer();
  pointerState.value = {
    id: event.pointerId,
    startX: event.clientX,
    startY: event.clientY,
    type: event.pointerType,
    didPreventDefault: shouldSuppressNativeLongPress,
    cancelTap: false,
  };
  capturePointer(event);
  didGesture.value = false;
  isDragging.value = false;

  longPressTimer = window.setTimeout(() => {
    didGesture.value = true;
    pointerState.value = null;
    suppressContextMenuUntil = Date.now() + CONTEXT_MENU_SUPPRESS_MS;
    resetDrag();
    openMenu({ deferHaptic: true });
  }, LONG_PRESS_MS);
}

function movePointer(event) {
  const state = pointerState.value;
  if (!state || state.id !== event.pointerId) return;

  const dx = event.clientX - state.startX;
  const dy = event.clientY - state.startY;
  const absX = Math.abs(dx);
  const absY = Math.abs(dy);

  if (absY > TAP_SLOP && absY > absX) {
    state.cancelTap = true;
    clearLongPressTimer();
    return;
  }

  if (absX > TAP_SLOP) {
    state.cancelTap = true;
    clearLongPressTimer();
  }

  if (dx <= 0 || absY > absX) return;

  isDragging.value = true;
  didGesture.value = true;
  dragOffset.value = Math.min(SWIPE_REVEAL_WIDTH, dx);
}

function finishPointer(event) {
  const state = pointerState.value;
  if (!state || state.id !== event.pointerId) return;

  clearLongPressTimer();
  releasePointer(event);
  pointerState.value = null;

  const shouldOpenFromPointer =
    state.didPreventDefault && !state.cancelTap && !didGesture.value;

  if (dragOffset.value >= SWIPE_REVEAL_WIDTH) {
    emit('complete', props.item);
  }

  resetDrag();

  if (shouldOpenFromPointer) {
    didGesture.value = true;
    emit('open', props.item);
  }
}

function cancelPointer(event) {
  const state = pointerState.value;
  if (state && event.pointerId !== state.id) return;
  clearLongPressTimer();
  releasePointer(event);
  pointerState.value = null;
  resetDrag();
}

function cancelMousePointer(event) {
  if (event.pointerType !== 'mouse' || !isDragging.value) return;
  cancelPointer(event);
}

function handleClick() {
  if (didGesture.value) {
    didGesture.value = false;
    return;
  }
  emit('open', props.item);
}

function handleContextMenu() {
  if (Date.now() < suppressContextMenuUntil) {
    clearHapticFallbackTimer();
    return;
  }

  openMenu({ withHaptic: false });
}

function openMenu({ withHaptic = true, deferHaptic = false } = {}) {
  clearLongPressTimer();
  clearDocumentSelection();
  if (withHaptic) {
    if (deferHaptic) {
      scheduleHapticFallback();
    } else {
      clearHapticFallbackTimer();
      telegram.triggerHapticImpact();
    }
  }
  emit('menu', props.item);
}

function resetDrag() {
  dragOffset.value = 0;
  isDragging.value = false;
}

function clearLongPressTimer() {
  if (!longPressTimer) return;
  window.clearTimeout(longPressTimer);
  longPressTimer = null;
}

function scheduleHapticFallback() {
  clearHapticFallbackTimer();
  hapticFallbackTimer = window.setTimeout(() => {
    hapticFallbackTimer = null;
    telegram.triggerHapticImpact();
  }, HAPTIC_FALLBACK_DELAY_MS);
}

function clearHapticFallbackTimer() {
  if (!hapticFallbackTimer) return;
  window.clearTimeout(hapticFallbackTimer);
  hapticFallbackTimer = null;
}

function clearDocumentSelection() {
  const selection = document.getSelection?.();
  if (!selection?.rangeCount) return;

  selection.removeAllRanges();
}

function capturePointer(event) {
  try {
    event.currentTarget.setPointerCapture?.(event.pointerId);
  } catch {
    // Pointer capture is an enhancement; the gesture still works without it.
  }
}

function releasePointer(event) {
  try {
    event.currentTarget.releasePointerCapture?.(event.pointerId);
  } catch {
    // Browsers may release capture automatically on cancel.
  }
}
</script>

<style scoped lang="scss">
.task-card {
  position: relative;
  width: 100%;
  min-height: 108px;
  overflow: hidden;
  color: $annonce-color-ink;
  border-radius: 14px;
  transition: transform $time-normal $ease-out;
  animation: mini-rise-in 220ms $ease-out both;
  animation-delay: var(--motion-delay, 0ms);
  will-change: transform;

  &:active {
    transform: scale(0.992);
  }

  &__complete-cue {
    position: absolute;
    inset: 0;
    font-size: 12px;
    font-weight: 700;
    line-height: 16px;
    color: #2e8c5a;
    background-color: #dcefe3;
    border-radius: 14px;
    opacity: 0.72;
    transition: opacity $time-normal $ease-out;

    :deep(svg) {
      position: absolute;
      top: 43px;
      left: 18px;
      width: 22px;
      height: 22px;
      transition: transform $time-normal $ease-out;
    }

    span {
      position: absolute;
      top: 35px;
      left: 54px;
      width: 90px;
    }
  }

  &__main {
    position: relative;
    z-index: 1;
    display: grid;
    gap: 6px;
    width: 100%;
    min-height: 108px;
    padding: 10px 12px;
    color: inherit;
    text-align: left;
    touch-action: pan-y;
    /* stylelint-disable-next-line property-no-vendor-prefix -- Needed for iOS WebView long-press selection. */
    -webkit-user-select: none;
    user-select: none;
    -webkit-touch-callout: none;
    background-color: #fefcf7;
    border: 1px solid #d4e2ea;
    border-radius: 14px;
    box-shadow: 0 2px 10px rgba(#0b304b, 0.06);
    transition:
      transform $time-normal $ease-out,
      border-color $time-fast $ease-out,
      box-shadow $time-fast $ease-out;
  }

  &--dragging &__main {
    transition: none;
  }

  &--ready &__main {
    border-color: rgba(#2e8c5a, 0.44);
    box-shadow: 0 8px 18px rgba(#2e8c5a, 0.13);
  }

  &--ready &__complete-cue {
    opacity: 1;

    :deep(svg) {
      transform: scale(1.01);
    }
  }

  &__meta,
  &__bottom {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: space-between;
    min-width: 0;
  }

  &__meta {
    font-size: 13px;
    font-weight: 700;
    line-height: 18px;
    color: #52636d;

    strong {
      max-width: 54%;
      overflow: hidden;
      text-overflow: ellipsis;
      font-weight: 500;
      text-align: right;
      white-space: nowrap;
    }
  }

  &__title {
    display: -webkit-box;
    overflow: hidden;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    font-size: 16px;
    font-weight: 600;
    line-height: 21px;
  }

  &__board {
    display: flex;
    gap: 6px;
    align-items: center;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
    font-weight: 500;
    line-height: 18px;
    color: #52636d;
    white-space: nowrap;

    i {
      flex: 0 0 6px;
      width: 6px;
      height: 6px;
      border-radius: 2px;
    }
  }

  &__avatars {
    display: flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: flex-end;
  }

  &__avatar {
    margin-left: -4px;

    &:first-child {
      margin-left: 0;
    }
  }

  &--dragging {
    transform: none;
  }
}

@media (hover: hover) {
  .task-card:hover:not(.task-card--dragging) {
    transform: translate3d(0, -1px, 0);
  }

  .task-card:hover:not(.task-card--dragging) .task-card__main {
    border-color: rgba($annonce-color-blue, 0.26);
    box-shadow: 0 12px 26px rgba(#0b304b, 0.1);
  }
}

@media (prefers-reduced-motion: reduce) {
  .task-card {
    transition-duration: 1ms;
    animation: none;
  }

  .task-card__complete-cue,
  .task-card__complete-cue :deep(svg),
  .task-card__main {
    transition-duration: 1ms;
  }
}
</style>
