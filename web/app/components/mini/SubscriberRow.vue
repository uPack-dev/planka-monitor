<template>
  <article
    class="subscriber-row"
    :class="{ 'subscriber-row--blocked': item.isBlocked }"
    @contextmenu.prevent="handleContextMenu"
    @pointerdown="startPointer"
    @pointermove="movePointer"
    @pointerup="finishPointer"
    @pointercancel="cancelPointer"
    @pointerleave="cancelMousePointer"
  >
    <MiniAvatar class="subscriber-row__avatar" :user="item" />
    <div class="subscriber-row__body">
      <strong>{{ telegramLabel }}</strong>
      <small>{{ workspaceLabel }}</small>
    </div>
    <span
      class="subscriber-row__status"
      :class="`subscriber-row__status--${statusTone}`"
    >
      {{ statusLabel }}
    </span>
    <button
      class="subscriber-row__more"
      type="button"
      aria-label="Дії з підписником"
      @pointerdown.stop
      @click.stop="openMenu"
    >
      <CIcon name="more-mini" />
    </button>
  </article>
</template>

<script setup>
const props = defineProps({
  item: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(['block', 'admin', 'menu']);
const telegram = useTelegramWebApp();

const LONG_PRESS_MS = 420;
const TAP_SLOP = 8;
const CONTEXT_MENU_SUPPRESS_MS = 800;
const HAPTIC_FALLBACK_DELAY_MS = 120;

const pointerState = shallowRef(null);
let longPressTimer = null;
let hapticFallbackTimer = null;
let suppressContextMenuUntil = 0;

const telegramLabel = computed(() => {
  const username = String(props.item.telegramUsername || '').replace(/^@+/, '');
  return username ? `@${username}` : '@unknown';
});

const workspaceLabel = computed(() => {
  const username = String(props.item.workspaceUsername || '').trim();
  return username ? `→ ${username}` : '↳ без привʼязки';
});

const statusLabel = computed(() => {
  if (props.item.isBlocked) return 'Заблоковано';
  if (props.item.isAdmin) return 'Адмін';
  return 'Користувач';
});

const statusTone = computed(() => {
  if (props.item.isBlocked) return 'blocked';
  if (props.item.isAdmin) return 'admin';
  return 'user';
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
  };
  capturePointer(event);

  longPressTimer = window.setTimeout(() => {
    pointerState.value = null;
    suppressContextMenuUntil = Date.now() + CONTEXT_MENU_SUPPRESS_MS;
    openMenu({ deferHaptic: true });
  }, LONG_PRESS_MS);
}

function movePointer(event) {
  const state = pointerState.value;
  if (!state || state.id !== event.pointerId) return;

  const dx = event.clientX - state.startX;
  const dy = event.clientY - state.startY;

  if (Math.abs(dx) <= TAP_SLOP && Math.abs(dy) <= TAP_SLOP) return;

  clearLongPressTimer();
}

function finishPointer(event) {
  const state = pointerState.value;
  if (!state || state.id !== event.pointerId) return;

  clearLongPressTimer();
  releasePointer(event);
  pointerState.value = null;
}

function cancelPointer(event) {
  const state = pointerState.value;
  if (state && event.pointerId !== state.id) return;

  clearLongPressTimer();
  releasePointer(event);
  pointerState.value = null;
}

function cancelMousePointer(event) {
  if (event.pointerType !== 'mouse') return;
  cancelPointer(event);
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
.subscriber-row {
  --subscriber-status-font-size: 13px;
  --subscriber-status-padding: 0 12px;
  --subscriber-status-width: 105px;

  display: grid;
  grid-template-columns: 34px minmax(0, 1fr) auto 24px;
  gap: 10px;
  align-items: center;
  min-height: 64px;
  padding: 10px 10px 10px 11px;
  color: $annonce-color-ink;
  touch-action: pan-y;
  /* stylelint-disable-next-line property-no-vendor-prefix -- Needed for iOS WebView long-press selection. */
  -webkit-user-select: none;
  user-select: none;
  background-color: #fefcf7;
  border: 1px solid #d4e2ea;
  border-radius: 12px;
  transition:
    border-color $time-fast $ease-out,
    box-shadow $time-fast $ease-out,
    opacity $time-fast $ease-out,
    transform $time-fast $ease-out;
  animation: mini-rise-in 220ms $ease-out both;
  animation-delay: var(--motion-delay, 0ms);
  -webkit-touch-callout: none;

  @media (width <= 359px) {
    --subscriber-status-font-size: 12px;
    --subscriber-status-padding: 0 8px;
    --subscriber-status-width: auto;

    grid-template-columns: 34px minmax(0, 1fr) minmax(84px, 96px) 24px;
    gap: 8px;
  }

  &__avatar {
    --mini-avatar-size: 34px;
    --mini-avatar-font-size: 13px;

    transition: transform $time-fast $ease-out;
  }

  &__body {
    display: grid;
    gap: 2px;
    min-width: 0;

    strong,
    small {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    strong {
      font-size: 14px;
      font-weight: 700;
      line-height: 18px;
    }

    small {
      font-size: 13px;
      font-weight: 500;
      line-height: 17px;
      color: $annonce-color-ink-muted;
    }
  }

  &__status {
    display: grid;
    place-items: center;
    width: var(--subscriber-status-width);
    height: 24px;
    padding: var(--subscriber-status-padding);
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: var(--subscriber-status-font-size);
    font-weight: 700;
    line-height: 16px;
    white-space: nowrap;
    border: 1px solid;
    border-radius: 999px;
    transition: transform $time-fast $ease-out;
  }

  &__status--user {
    color: #2e8c5a;
    background-color: rgba(#2e8c5a, 0.1);
    border-color: rgba(#2e8c5a, 0.4);
  }

  &__status--admin {
    color: $annonce-color-blue;
    background-color: rgba($annonce-color-navy, 0.1);
    border-color: rgba($annonce-color-blue, 0.4);
  }

  &__status--blocked {
    --subscriber-status-width: 122px;

    color: #d94d67;
    background-color: rgba(#d94d67, 0.1);
    border-color: rgba(#d94d67, 0.4);

    @media (width <= 359px) {
      --subscriber-status-width: auto;
    }
  }

  &__more {
    display: grid;
    place-items: center;
    width: 24px;
    height: 36px;
    color: $annonce-color-blue;
    transition:
      background-color $time-fast $ease-out,
      transform $time-fast $ease-out;

    :deep(svg) {
      --fill-0: #275f88;

      width: 20px;
      height: 20px;
    }
  }

  &--blocked {
    opacity: 0.76;

    strong {
      color: $annonce-color-ink-muted;
    }
  }
}

@media (hover: hover) {
  .subscriber-row:hover {
    border-color: rgba($annonce-color-blue, 0.28);
    box-shadow: 0 12px 24px rgba(#0b304b, 0.09);
    transform: translate3d(0, -1px, 0);
  }

  .subscriber-row:hover .subscriber-row__avatar {
    transform: scale(1.02);
  }

  .subscriber-row__more:hover {
    background-color: rgba($annonce-color-blue, 0.08);
    transform: scale(1.02);
  }
}

@media (prefers-reduced-motion: reduce) {
  .subscriber-row {
    transition-duration: 1ms;
    animation: none;
  }

  .subscriber-row__avatar,
  .subscriber-row__status,
  .subscriber-row__more {
    transition-duration: 1ms;
  }
}
</style>
