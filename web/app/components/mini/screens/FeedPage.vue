<template>
  <section class="screen feed-screen">
    <div v-if="isLoading && !items.length" class="loader">
      Завантажуємо стрічку...
    </div>

    <MiniEmptyState
      v-else-if="!items.length"
      icon="bell"
      title="Тут тихо"
      text="Нові коментарі, призначення і виконання зʼявляться в цій стрічці."
      action-label="Оновити"
      @action="$emit('refresh')"
    />

    <div v-else class="feed-timeline">
      <section
        v-for="(group, groupIndex) in feedGroups"
        :key="group.key"
        class="feed-timeline__group"
        :class="`feed-timeline__group--${group.tone}`"
        :style="{ '--motion-delay': `${groupIndex * 24}ms` }"
      >
        <header class="feed-timeline__header">
          <span>{{ group.label }}</span>
        </header>

        <button
          v-for="(item, itemIndex) in group.items"
          :key="item.id"
          class="feed-card"
          :class="{ 'feed-card--unread': !item.isRead }"
          :style="{ '--motion-delay': `${itemIndex * 20}ms` }"
          type="button"
          @click="handleClick(item)"
          @contextmenu.prevent="handleContextMenu(item)"
          @pointerdown="startPointer($event, item)"
          @pointermove="movePointer"
          @pointerup="finishPointer"
          @pointercancel="cancelPointer"
        >
          <MiniAvatar class="feed-card__avatar" :user="item.actor" size="md" />

          <span class="feed-card__body">
            <span class="feed-card__top">
              <strong class="feed-card__event">{{ feedTitle(item) }}</strong>
              <time
                class="feed-card__time"
                :datetime="itemDatetime(item.createdAt)"
              >
                {{ itemTime(item) }}
              </time>
            </span>

            <strong v-if="item.cardTitle" class="feed-card__title">
              {{ item.cardTitle }}
            </strong>

            <span v-if="item.text" class="feed-card__preview">
              {{ item.text }}
            </span>

            <span class="feed-card__footer">
              <span class="feed-card__board">
                <i :style="{ backgroundColor: boardColor(item) }" />
                {{ boardName(item) }}
              </span>
              <span
                v-if="!item.isRead"
                class="feed-card__unread"
                aria-label="Непрочитано"
              />
            </span>
          </span>
        </button>
      </section>
    </div>
  </section>
</template>

<script setup>
import {
  dateKey,
  dayLabel,
  formatClockTime,
  parseValidDate,
} from '@/utils/miniDate';

const props = defineProps({
  items: {
    type: Array,
    required: true,
  },
  isLoading: {
    type: Boolean,
    required: true,
  },
  feedTitle: {
    type: Function,
    required: true,
  },
  relativeDate: {
    type: Function,
    required: true,
  },
});

const emit = defineEmits(['refresh', 'open', 'menu']);
const telegram = useTelegramWebApp();

const LONG_PRESS_MS = 420;
const TAP_SLOP = 8;
const CONTEXT_MENU_SUPPRESS_MS = 800;
const HAPTIC_FALLBACK_DELAY_MS = 120;

const pointerState = shallowRef(null);
const didGesture = shallowRef(false);
let longPressTimer = null;
let hapticFallbackTimer = null;
let suppressContextMenuUntil = 0;

const feedGroups = computed(() => {
  const groups = new Map();

  for (const item of props.items) {
    const createdAt = parseValidDate(item.createdAt);
    const key = createdAt ? dateKey(createdAt) : 'unknown';

    if (!groups.has(key)) {
      groups.set(key, {
        key,
        label: createdAt
          ? dayLabel(createdAt, { capitalized: true })
          : 'Без дати',
        tone: isToday(createdAt) ? 'today' : 'past',
        items: [],
      });
    }

    groups.get(key).items.push(item);
  }

  return Array.from(groups.values());
});

onBeforeUnmount(() => {
  clearLongPressTimer();
  clearHapticFallbackTimer();
});

function itemTime(item) {
  return formatClockTime(item.createdAt) || props.relativeDate(item.createdAt);
}

function itemDatetime(value) {
  return parseValidDate(value)?.toISOString() || undefined;
}

function boardName(item) {
  return item.boardName || 'Workspace';
}

function boardColor(item) {
  if (item.boardColor) return item.boardColor;

  const palette = ['#d94d67', '#2e8c5a', '#275f88', '#5995c1', '#ffc261'];
  const seed = Array.from(boardName(item)).reduce(
    (sum, char) => sum + char.charCodeAt(0),
    0,
  );

  return palette[seed % palette.length];
}

function isToday(value) {
  const date = parseValidDate(value);
  return Boolean(date && dateKey(date) === dateKey(new Date()));
}

function startPointer(event, item) {
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
    item,
    startX: event.clientX,
    startY: event.clientY,
    didPreventDefault: shouldSuppressNativeLongPress,
    cancelTap: false,
  };
  capturePointer(event);
  didGesture.value = false;

  longPressTimer = window.setTimeout(() => {
    didGesture.value = true;
    pointerState.value = null;
    suppressContextMenuUntil = Date.now() + CONTEXT_MENU_SUPPRESS_MS;
    openMenu(item, { deferHaptic: true });
  }, LONG_PRESS_MS);
}

function movePointer(event) {
  const state = pointerState.value;
  if (!state || state.id !== event.pointerId) return;

  const dx = event.clientX - state.startX;
  const dy = event.clientY - state.startY;

  if (Math.abs(dx) > TAP_SLOP || Math.abs(dy) > TAP_SLOP) {
    state.cancelTap = true;
    didGesture.value = true;
    clearLongPressTimer();
  }
}

function finishPointer(event) {
  const state = pointerState.value;
  if (!state || state.id !== event.pointerId) return;

  clearLongPressTimer();
  releasePointer(event);
  pointerState.value = null;

  const shouldOpenFromPointer =
    state.didPreventDefault && !state.cancelTap && !didGesture.value;

  if (shouldOpenFromPointer) {
    didGesture.value = true;
    emit('open', state.item);
  }
}

function cancelPointer(event) {
  const state = pointerState.value;
  if (state && event.pointerId !== state.id) return;
  clearLongPressTimer();
  releasePointer(event);
  pointerState.value = null;
}

function handleClick(item) {
  if (didGesture.value) {
    didGesture.value = false;
    return;
  }
  emit('open', item);
}

function handleContextMenu(item) {
  if (Date.now() < suppressContextMenuUntil) {
    clearHapticFallbackTimer();
    return;
  }

  openMenu(item, { withHaptic: false });
}

function openMenu(item, { withHaptic = true, deferHaptic = false } = {}) {
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
  emit('menu', item);
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
    // Pointer capture is optional for the gesture.
  }
}

function releasePointer(event) {
  try {
    event.currentTarget.releasePointerCapture?.(event.pointerId);
  } catch {
    // Browsers may release capture automatically.
  }
}
</script>

<style scoped lang="scss">
.screen {
  display: grid;
  gap: 14px;
  align-content: start;
  min-height: max-content;
}

.loader {
  padding: 26px;
  font-size: 14px;
  font-weight: 700;
  color: $annonce-color-ink-muted;
  text-align: center;
}

.feed-screen .loader {
  display: grid;
  place-items: center;
  min-height: 440px;
}

.feed-timeline {
  position: relative;
  display: grid;
  gap: 14px;
  padding: 2px 0 0 14px;
}

.feed-timeline__group {
  --timeline-color: #184566;
  --timeline-text-color: #ffffff;

  position: relative;
  display: grid;
  gap: 14px;
  animation: mini-rise-in 220ms $ease-out both;
  animation-delay: var(--motion-delay, 0ms);

  &::before {
    position: absolute;
    top: 0;
    bottom: -14px;
    left: -14px;
    width: 4px;
    content: '';
    background-color: var(--timeline-color);
    border-radius: 2px;
  }

  &:last-child::before {
    bottom: 14px;
  }

  &::after {
    position: absolute;
    top: 14px;
    right: -16px;
    left: -10px;
    height: 2px;
    content: '';
    background-color: var(--timeline-color);
  }

  &--today {
    --timeline-color: #ffc261;
    --timeline-text-color: #082d3d;
  }
}

.feed-timeline__header {
  position: relative;
  z-index: 1;
  display: flex;

  span {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    min-height: 28px;
    padding: 0 12px;
    overflow: hidden;
    font-size: 13px;
    font-weight: 700;
    line-height: 16px;
    color: var(--timeline-text-color);
    white-space: nowrap;
    background-color: var(--timeline-color);
    border-radius: 14px;
  }

  span::before {
    width: 4px;
    height: 4px;
    content: '';
    background-color: currentcolor;
    border-radius: 50%;
  }
}

.feed-card {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr);
  gap: 12px;
  align-items: start;
  width: 100%;
  min-height: 88px;
  padding: 10px;
  color: $annonce-color-ink;
  text-align: left;
  touch-action: pan-y;
  /* stylelint-disable-next-line property-no-vendor-prefix -- Needed for iOS WebView long-press selection. */
  -webkit-user-select: none;
  user-select: none;
  background-color: #fffdf8;
  border: 1.2px solid #d4e2ea;
  border-radius: 14px;
  box-shadow: 0 2px 10px rgba(#0b304b, 0.08);
  transition:
    transform $time-fast $ease-out,
    border-color $time-fast $ease-out,
    box-shadow $time-fast $ease-out;
  animation: mini-rise-in 220ms $ease-out both;
  animation-delay: var(--motion-delay, 0ms);
  -webkit-touch-callout: none;

  &:active {
    border-color: rgba(#184566, 0.22);
    box-shadow: 0 1px 6px rgba(#0b304b, 0.08);
    transform: scale(0.992);
  }

  &__avatar {
    margin-top: 0;
    transition: transform $time-normal $ease-out;
  }

  &__body {
    display: grid;
    align-content: start;
    min-width: 0;
  }

  &__top {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 10px;
    align-items: start;
    min-width: 0;
  }

  &__event {
    min-width: 0;
    max-width: 212px;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 14.5px;
    font-weight: 500;
    line-height: 20px;
    color: #182a33;
  }

  &__time {
    flex: 0 0 auto;
    max-width: 64px;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
    font-weight: 700;
    line-height: 17px;
    color: #52636d;
    white-space: nowrap;
  }

  &__title {
    min-width: 0;
    margin-top: 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
    font-weight: 700;
    line-height: 17px;
    color: #182a33;
  }

  &__preview {
    display: -webkit-box;
    padding: 10px;
    margin-top: 10px;
    overflow: hidden;
    -webkit-line-clamp: 2;
    font-size: 12px;
    font-weight: 500;
    line-height: 17px;
    color: #52636d;
    background-color: #e1edf5;
    border-radius: 8px;
    -webkit-box-orient: vertical;
  }

  &__footer {
    display: flex;
    gap: 10px;
    align-items: center;
    justify-content: space-between;
    min-height: 18px;
    margin-top: 10px;
  }

  &__board {
    display: inline-flex;
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

  &__unread {
    flex: 0 0 8px;
    width: 8px;
    height: 8px;
    background-color: $annonce-color-yellow-action;
    border-radius: 50%;
  }
}

@media (hover: hover) {
  .feed-card:hover {
    border-color: rgba($annonce-color-blue, 0.24);
    box-shadow: 0 14px 30px rgba(#0b304b, 0.12);
    transform: translate3d(0, -1px, 0);
  }

  .feed-card:hover .feed-card__avatar {
    transform: scale(1.02);
  }
}

@media (prefers-reduced-motion: reduce) {
  .feed-timeline__group,
  .feed-card,
  .feed-card__unread {
    animation: none;
  }

  .feed-card,
  .feed-card__avatar {
    transition-duration: 1ms;
  }
}
</style>
