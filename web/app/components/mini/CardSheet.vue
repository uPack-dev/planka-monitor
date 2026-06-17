<template>
  <Teleport to="body">
    <div v-if="card" class="card-detail">
      <section class="card-detail__panel" role="dialog" aria-modal="true">
        <header class="card-detail__header">
          <div class="card-detail__header-top">
            <button
              class="card-detail__nav-button card-detail__nav-button--back"
              type="button"
              aria-label="Назад"
              @click="$emit('close')"
            >
              <CIcon name="back-mini" />
            </button>
            <span class="card-detail__eyebrow">{{ card.boardName }}</span>
            <button
              class="card-detail__nav-button"
              type="button"
              aria-label="Додаткові дії"
              disabled
            >
              <CIcon name="more-mini" />
            </button>
          </div>

          <h2>{{ card.title }}</h2>
          <p class="card-detail__due">
            <span>{{ dueLabel }}</span>
            <strong>{{ relativeDueLabel }}</strong>
          </p>
        </header>

        <div class="card-detail__body">
          <section
            v-if="card.members?.length"
            class="card-detail__section card-detail__section--assignees"
          >
            <h3>Призначені:</h3>
            <div class="card-detail__assignees">
              <span
                v-for="(member, memberIndex) in card.members"
                :key="member.id || member.username"
                class="card-detail__assignee"
                :style="{ '--motion-delay': `${memberIndex * 16}ms` }"
              >
                <MiniAvatar
                  class="card-detail__avatar--assignee"
                  :user="member"
                  size="sm"
                />
                <span>{{ displayUserName(member) }}</span>
              </span>
            </div>
          </section>

          <section
            v-if="card.tasks?.length"
            class="card-detail__section card-detail__section--tasks"
            aria-label="Чек-ліст"
          >
            <header class="card-detail__section-header">
              <h3>Чек-ліст:</h3>
              <span>{{ completedTaskCount }}/{{ card.tasks.length }}</span>
            </header>
            <div class="card-detail__progress" aria-hidden="true">
              <span :style="{ width: taskProgressWidth }" />
            </div>
            <div class="card-detail__tasks">
              <button
                v-for="(task, taskIndex) in card.tasks"
                :key="task.id"
                class="card-detail__task-row"
                :style="{ '--motion-delay': `${taskIndex * 16}ms` }"
                type="button"
                :disabled="task.isCompleted"
                @click="$emit('complete-task', task.id)"
              >
                <span
                  class="card-detail__task-check"
                  :class="{ 'is-done': task.isCompleted }"
                >
                  <CIcon v-if="task.isCompleted" name="check" />
                </span>
                <span class="card-detail__task-title">{{ task.name }}</span>
                <MiniAvatar
                  v-if="task.assignee"
                  :user="task.assignee"
                  size="xs"
                />
              </button>
            </div>
          </section>

          <section v-if="card.description" class="card-detail__section">
            <h3>Опис:</h3>
            <div
              class="card-detail__rich-text"
              v-html="formatRichText(card.description)"
            />
          </section>

          <section class="card-detail__section card-detail__section--comments">
            <h3>Коментарі:</h3>
            <div v-if="card.comments?.length" class="card-detail__comments">
              <div
                v-for="(entry, commentIndex) in card.comments"
                :key="entry.id"
                class="card-detail__comment-row"
                :class="{
                  'card-detail__comment-row--own': isOwnComment(entry),
                }"
                :style="{ '--motion-delay': `${commentIndex * 18}ms` }"
              >
                <MiniAvatar
                  v-if="!isOwnComment(entry)"
                  :user="entry.author"
                  size="md"
                />
                <article class="card-detail__comment-bubble">
                  <header>
                    <strong>{{ commentAuthorName(entry) }}</strong>
                    <time :datetime="entry.createdAt">
                      {{ commentTime(entry.createdAt) }}
                    </time>
                  </header>
                  <div
                    class="card-detail__rich-text"
                    :class="{
                      'card-detail__rich-text--own': isOwnComment(entry),
                    }"
                    v-html="formatRichText(entry.text)"
                  />
                </article>
              </div>
            </div>
            <p v-else class="card-detail__empty">Коментарів ще немає.</p>
          </section>
        </div>

        <div
          v-if="isComposerVisible"
          class="card-detail__composer-backdrop"
          :class="{
            'card-detail__composer-backdrop--open': isComposerOpen,
            'card-detail__composer-backdrop--dragging': isComposerDragging,
          }"
          @click.self="closeCommentComposer"
        >
          <form
            class="card-detail__composer"
            :style="{
              transform: `translate3d(0, ${composerPanelOffset}px, 0)`,
            }"
            @submit.prevent="sendComment"
          >
            <div
              class="card-detail__composer-drag-zone"
              @pointerdown="startComposerDrag"
              @pointermove="moveComposerDrag"
              @pointerup="finishComposerDrag"
              @pointercancel="cancelComposerDrag"
            >
              <span class="card-detail__composer-handle" aria-hidden="true" />
            </div>
            <p>Відповісти на</p>
            <strong>{{ card.title }}</strong>
            <textarea
              ref="commentInput"
              v-model="commentText"
              rows="4"
              placeholder="Напишіть коментар..."
            />
            <div class="card-detail__composer-actions">
              <button class="card-detail__mention-button" type="button">
                <CIcon name="mention-mini" />
                Згадати
              </button>
              <button type="submit" :disabled="!commentText.trim()">
                Надіслати
              </button>
            </div>
          </form>
        </div>

        <footer class="card-detail__actions">
          <button
            class="card-detail__action card-detail__action--primary"
            type="button"
            :disabled="primaryActionDisabled"
            @click="completePrimaryAction"
          >
            <CIcon name="check-filled-mini" />
            <span>Позначити виконаною</span>
          </button>
          <button
            class="card-detail__action"
            :class="{ 'card-detail__action--active': isComposerOpen }"
            type="button"
            @click="toggleCommentComposer"
          >
            <CIcon name="message-filled-mini" />
            <span>Коментар</span>
          </button>
          <a
            v-if="workspaceUrl"
            class="card-detail__action card-detail__action--icon"
            :href="workspaceUrl"
            target="_blank"
            rel="noreferrer"
            aria-label="Відкрити в Workspace"
          >
            <CIcon name="workspace-mini" />
          </a>
        </footer>

        <MiniBottomNav active="tasks" @update:active="handleNavigate" />
      </section>
    </div>
  </Teleport>
</template>

<script setup>
import {
  dayLabel,
  formatClockTime,
  parseValidDate,
  relativeDeadlineLabel,
} from '@/utils/miniDate';
import { displayUserName, normalizeUsername } from '@/utils/miniUsers';
import { formatRichText as formatMiniRichText } from '@/utils/richText';

const props = defineProps({
  card: {
    type: Object,
    default: null,
  },
  workspaceUrl: {
    type: String,
    default: '',
  },
  currentUsername: {
    type: String,
    default: '',
  },
  composerRequestKey: {
    type: Number,
    default: 0,
  },
});

const emit = defineEmits([
  'close',
  'complete-card',
  'complete-task',
  'comment',
  'navigate',
]);

const commentText = shallowRef('');
const commentInput = ref(null);
const isComposerVisible = shallowRef(false);
const isComposerOpen = shallowRef(false);
const isComposerDragging = shallowRef(false);
const composerDragOffset = shallowRef(0);
const composerDragState = shallowRef(null);
let composerCloseTimer = null;
let composerFocusTimer = null;

const COMPOSER_TRANSITION_MS = 260;
const COMPOSER_CLOSE_THRESHOLD = 88;
const COMPOSER_HIDDEN_OFFSET = 430;

watch(
  () => props.card?.id,
  () => {
    commentText.value = '';
    resetCommentComposer();
  },
);

watch(
  () => props.composerRequestKey,
  () => {
    if (props.card?.id) {
      openCommentComposer();
    }
  },
);

onMounted(() => {
  document.addEventListener('keydown', handleKeydown);
});

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown);
  clearTimeout(composerCloseTimer);
  clearTimeout(composerFocusTimer);
});

const dueDate = computed(() => parseValidDate(props.card?.dueAt));

const dueLabel = computed(() => {
  if (!dueDate.value) return 'Без дедлайну';
  return `Дедлайн: ${dayLabel(dueDate.value)} ${formatClockTime(
    dueDate.value,
  )}`;
});

const relativeDueLabel = computed(() => {
  if (!dueDate.value) return '';
  return relativeDeadlineLabel(dueDate.value);
});

const completedTaskCount = computed(
  () => props.card?.tasks?.filter((task) => task.isCompleted).length || 0,
);

const taskProgressWidth = computed(() => {
  const total = props.card?.tasks?.length || 0;
  if (!total) return '0%';
  return `${Math.round((completedTaskCount.value / total) * 100)}%`;
});

const firstIncompleteTask = computed(
  () => props.card?.tasks?.find((task) => !task.isCompleted) || null,
);

const primaryActionDisabled = computed(
  () =>
    !props.card || (!firstIncompleteTask.value && props.card.isDueCompleted),
);

const normalizedCurrentUsername = computed(() =>
  normalizeUsername(props.currentUsername),
);
const composerPanelOffset = computed(() =>
  isComposerOpen.value ? composerDragOffset.value : COMPOSER_HIDDEN_OFFSET,
);

function completePrimaryAction() {
  if (!props.card || primaryActionDisabled.value) return;
  if (firstIncompleteTask.value) {
    emit('complete-task', firstIncompleteTask.value.id);
    return;
  }
  emit('complete-card', props.card.id);
}

function handleNavigate(tab) {
  if (tab === 'tasks') return;
  emit('navigate', tab);
}

function handleKeydown(event) {
  if (!props.card || event.key !== 'Escape') return;
  event.preventDefault();
  if (isComposerVisible.value) {
    closeCommentComposer();
    return;
  }
  emit('close');
}

async function openCommentComposer() {
  clearTimeout(composerCloseTimer);
  clearTimeout(composerFocusTimer);
  isComposerVisible.value = true;
  isComposerDragging.value = false;
  composerDragOffset.value = 0;
  await nextTick();
  const openComposer = () => {
    isComposerOpen.value = true;
    scheduleCommentFocus();
  };

  if (typeof requestAnimationFrame === 'function') {
    requestAnimationFrame(openComposer);
  } else {
    openComposer();
  }
}

function toggleCommentComposer() {
  if (isComposerOpen.value) {
    closeCommentComposer();
    return;
  }

  openCommentComposer();
}

function closeCommentComposer() {
  clearTimeout(composerCloseTimer);
  clearTimeout(composerFocusTimer);
  blurCommentInput();
  isComposerDragging.value = false;
  composerDragState.value = null;
  composerDragOffset.value = 0;
  isComposerOpen.value = false;
  composerCloseTimer = setTimeout(() => {
    isComposerVisible.value = false;
  }, COMPOSER_TRANSITION_MS);
}

function resetCommentComposer() {
  clearTimeout(composerCloseTimer);
  clearTimeout(composerFocusTimer);
  blurCommentInput();
  isComposerVisible.value = false;
  isComposerOpen.value = false;
  isComposerDragging.value = false;
  composerDragState.value = null;
  composerDragOffset.value = 0;
}

function scheduleCommentFocus() {
  clearTimeout(composerFocusTimer);
  composerFocusTimer = setTimeout(() => {
    if (!isComposerVisible.value || !isComposerOpen.value) return;
    focusCommentInput();
  }, COMPOSER_TRANSITION_MS);
}

function focusCommentInput() {
  const input = commentInput.value;
  if (!input) return;

  try {
    input.focus({ preventScroll: true });
  } catch {
    input.focus();
  }
}

function blurCommentInput() {
  const input = commentInput.value;
  if (!input || document.activeElement !== input) return;
  input.blur();
}

function sendComment() {
  const text = commentText.value.trim();
  if (!text) return;
  emit('comment', text);
  commentText.value = '';
  closeCommentComposer();
}

function startComposerDrag(event) {
  if (event.button && event.button !== 0) return;
  isComposerDragging.value = true;
  composerDragState.value = {
    id: event.pointerId,
    startY: event.clientY,
  };
  capturePointer(event);
}

function moveComposerDrag(event) {
  const state = composerDragState.value;
  if (!state || state.id !== event.pointerId) return;

  const dy = event.clientY - state.startY;
  composerDragOffset.value = Math.max(0, dy);
}

function finishComposerDrag(event) {
  const state = composerDragState.value;
  if (!state || state.id !== event.pointerId) return;

  releasePointer(event);
  composerDragState.value = null;
  isComposerDragging.value = false;

  if (composerDragOffset.value >= COMPOSER_CLOSE_THRESHOLD) {
    closeCommentComposer();
    return;
  }

  composerDragOffset.value = 0;
}

function cancelComposerDrag(event) {
  const state = composerDragState.value;
  if (state && state.id !== event.pointerId) return;
  releasePointer(event);
  composerDragState.value = null;
  isComposerDragging.value = false;
  composerDragOffset.value = 0;
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

function isOwnComment(entry) {
  const author = normalizeUsername(entry?.author?.username);
  return Boolean(author && author === normalizedCurrentUsername.value);
}

function commentAuthorName(entry) {
  return entry?.author?.name || entry?.author?.username || 'Workspace';
}

function commentTime(value) {
  return formatClockTime(value);
}

function formatRichText(value) {
  return formatMiniRichText(value, {
    mentionClass: 'card-detail__mention',
  });
}
</script>

<style scoped lang="scss">
.card-detail {
  position: fixed;
  inset: 0;
  z-index: 320;
  display: block;
  background-color: $annonce-color-cream;
  animation: card-detail-fade 220ms $ease-out both;

  &__panel {
    position: relative;
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100vh;
    height: 100dvh;
    height: var(--tg-viewport-height, 100dvh);
    overflow: hidden;
    color: $annonce-color-ink;
    background-color: $annonce-color-cream;
  }

  &__header {
    flex: 0 0 auto;
    padding: calc(var(--tg-safe-top, 0px) + 20px) 16px 15px;
    color: $annonce-color-text;
    background-color: $annonce-color-navy;
    animation: mini-rise-in 200ms $ease-out both;
  }

  &__header-top {
    display: grid;
    grid-template-columns: 44px minmax(0, 1fr) 44px;
    gap: 12px;
    align-items: center;
    margin-bottom: 12px;
  }

  &__nav-button {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    color: $annonce-color-text;
    background-color: rgba(#ffffff, 0.16);
    border: 1px solid rgba(#ffffff, 0.26);
    border-radius: 14px;
    transition:
      background-color $time-fast $ease-out,
      border-color $time-fast $ease-out,
      transform $time-fast $ease-out;

    :deep(svg) {
      width: 20px;
      height: 20px;
    }
  }

  &__nav-button--back {
    :deep(svg) {
      width: 16px;
      height: 16px;
    }
  }

  &__eyebrow {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 12px;
    font-weight: 800;
    line-height: 16px;
    color: rgba($annonce-color-muted-strong, 0.66);
    text-align: center;
    white-space: nowrap;
  }

  &__due {
    display: flex;
    gap: 10px;
    align-items: center;
    justify-content: space-between;
    margin-top: 10px;
    font-size: 13px;
    font-weight: 700;
    line-height: 16px;
    color: rgba(#ffffff, 0.7);

    span,
    strong {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    strong {
      font-weight: 700;
      text-align: right;
    }
  }

  &__body {
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
    gap: 18px;
    min-height: 0;
    padding: 18px 16px calc(var(--tg-safe-bottom, 0px) + 154px);
    overflow-y: auto;
    scroll-behavior: smooth;
    overscroll-behavior: contain;
    scrollbar-width: none;
    -webkit-overflow-scrolling: touch;

    &::-webkit-scrollbar {
      display: none;
    }
  }

  &__section {
    display: grid;
    flex-shrink: 0;
    gap: 8px;
    min-width: 0;
    animation: mini-rise-in 210ms $ease-out both;
  }

  &__section:nth-child(2) {
    animation-delay: 16ms;
  }

  &__section:nth-child(3) {
    animation-delay: 32ms;
  }

  &__section:nth-child(4) {
    animation-delay: 48ms;
  }

  &__section-header {
    display: flex;
    gap: 10px;
    align-items: center;
    justify-content: space-between;

    span {
      flex: 0 0 auto;
      font-size: 13px;
      font-weight: 800;
      line-height: 20px;
      color: $annonce-color-ink;
    }
  }

  &__assignees {
    display: flex;
    gap: 10px;
    min-width: 0;
    overflow-x: auto;
    scrollbar-width: none;

    &::-webkit-scrollbar {
      display: none;
    }
  }

  &__assignee {
    display: inline-flex;
    flex: 0 0 auto;
    gap: 10px;
    align-items: center;
    min-width: 0;
    font-size: 13px;
    font-weight: 500;
    line-height: 18px;
    color: $annonce-color-ink;
    white-space: nowrap;
    animation: mini-rise-in 180ms $ease-out both;
    animation-delay: var(--motion-delay, 0ms);
  }

  &__progress {
    height: 6px;
    overflow: hidden;
    background-color: #d4e2ea;
    border-radius: 999px;

    span {
      display: block;
      height: 100%;
      background-color: $annonce-color-yellow;
      border-radius: inherit;
      transition: width $time-normal $ease-out;
    }
  }

  &__tasks {
    display: grid;
    gap: 6px;
  }

  &__task-row {
    display: grid;
    grid-template-columns: 28px minmax(0, 1fr) 18px;
    gap: 10px;
    align-items: center;
    min-height: 38px;
    padding: 8px 10px 8px 18px;
    overflow: hidden;
    color: $annonce-color-ink;
    text-align: left;
    background-color: #ffffff;
    border: 1.5px solid $annonce-color-stroke;
    border-radius: 12px;
    transition:
      border-color $time-fast $ease-out,
      box-shadow $time-fast $ease-out,
      transform $time-fast $ease-out;
    animation: mini-rise-in 190ms $ease-out both;
    animation-delay: var(--motion-delay, 0ms);

    &:disabled {
      opacity: 0.72;
    }
  }

  &__task-check {
    display: grid;
    place-items: center;
    width: 18px;
    height: 18px;
    color: #ffffff;
    border: 1.5px solid $annonce-color-blue;
    border-radius: 50%;
    transition:
      background-color $time-fast $ease-out,
      border-color $time-fast $ease-out,
      transform $time-fast $ease-out;

    &.is-done {
      background-color: #2e8c5a;
      border-color: #2e8c5a;
    }

    :deep(svg) {
      width: 12px;
      height: 12px;
    }
  }

  &__task-title {
    display: -webkit-box;
    min-width: 0;
    overflow: hidden;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    font-size: 14px;
    font-weight: 800;
    line-height: 18px;
  }

  &__rich-text {
    display: block;
    min-width: 0;
    font-size: 13px;
    line-height: 1.25;
    overflow-wrap: anywhere;
    white-space: normal;

    :deep(*) {
      max-width: 100%;
      line-height: inherit;
    }

    :deep(> * + *) {
      margin-top: 8px;
    }

    :deep(p) {
      margin: 0;
    }

    :deep(h4) {
      margin: 0;
      font-size: 14px;
      font-weight: 800;
    }

    :deep(ol),
    :deep(ul) {
      display: flex;
      flex-direction: column;
      gap: 6px;
      padding-left: 20px;
      margin: 0;
      list-style-position: outside;
    }

    :deep(ol) {
      list-style-type: decimal;
    }

    :deep(ul) {
      list-style-type: disc;
    }

    :deep(li) {
      display: list-item;
      padding-left: 2px;
    }

    :deep(a) {
      font-weight: 700;
      color: $annonce-color-blue;
      text-decoration: underline;
      text-decoration-thickness: 1px;
      text-underline-offset: 3px;
    }

    :deep(strong) {
      font-weight: 800;
    }

    :deep(code) {
      padding: 1px 4px;
      font-size: 12px;
      color: $annonce-color-blue;
      background-color: #e8f1f6;
      border-radius: 4px;
    }

    :deep(.card-detail__mention) {
      font-weight: 800;
      color: $annonce-color-blue;
    }
  }

  &__rich-text--own {
    color: #ffffff;

    :deep(a),
    :deep(.card-detail__mention) {
      color: #ffffff;
    }

    :deep(code) {
      color: #ffffff;
      background-color: rgba(#ffffff, 0.16);
    }
  }

  &__comments {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  &__comment-row {
    display: flex;
    gap: 10px;
    align-items: start;
    justify-content: start;
    width: 100%;
    animation: mini-rise-in 200ms $ease-out both;
    animation-delay: var(--motion-delay, 0ms);
  }

  &__comment-row--own {
    justify-content: end;

    .card-detail__comment-bubble {
      color: #ffffff;
      background-color: $annonce-color-blue;
      border-color: $annonce-color-blue;
      border-top-left-radius: 14px;
      border-top-right-radius: 2px;
    }

    time,
    strong {
      color: rgba(#ffffff, 0.68);
    }
  }

  &__comment-bubble {
    flex: 0 1 300px;
    min-width: 0;
    max-width: 100%;
    padding: 8px 10px;
    overflow-wrap: anywhere;
    background-color: #ffffff;
    border: 1px solid #c9dde8;
    border-radius: 2px 14px 14px;
    box-shadow: 0 6px 18px rgba(#0b304b, 0.06);

    header {
      display: grid;
      grid-template-columns: minmax(0, 1fr) auto;
      gap: 8px;
      align-items: center;
      margin-bottom: 4px;
    }

    strong {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 14.5px;
      font-weight: 800;
      line-height: 20px;
      color: $annonce-color-ink;
      white-space: nowrap;
    }

    time {
      font-size: 12px;
      font-weight: 600;
      color: $annonce-color-ink-muted;
    }
  }

  &__avatar--assignee {
    flex: 0 0 26px;
  }

  &__empty {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    color: $annonce-color-ink-muted;
  }

  &__composer-backdrop {
    position: absolute;
    inset: 0 0 calc(var(--tg-safe-bottom, 0px) + 144px);
    z-index: 4;
    display: flex;
    align-items: end;
    pointer-events: none;
    background-color: rgba(#0b304b, 0);
    transition: background-color $time-normal $ease-out;
  }

  &__composer-backdrop--open {
    pointer-events: auto;
    background-color: rgba(#0b304b, 0.45);
  }

  &__composer {
    position: relative;
    display: grid;
    gap: 12px;
    width: 100%;
    padding: 48px 18px 20px;
    background-color: #ffffff;
    border-radius: 24px 24px 0 0;
    box-shadow: 0 -22px 58px -28px rgba(#0b304b, 0.46);
    transition:
      box-shadow $time-normal $ease-out,
      transform $time-normal $ease-out;
    will-change: transform;

    p {
      margin: 0;
      font-size: 12px;
      font-weight: 800;
      line-height: 16px;
      color: $annonce-color-ink;
    }

    strong {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      font-family: $annonce-font-brand;
      font-size: 20px;
      font-weight: 400;
      line-height: 24px;
      color: $annonce-color-blue;
      white-space: nowrap;
    }

    textarea {
      width: 100%;
      min-height: 112px;
      padding: 10px;
      font-size: 14.5px;
      font-weight: 700;
      line-height: 20px;
      color: $annonce-color-ink;
      resize: none;
      background-color: #e9eff3;
      border: 1.5px solid #d4e2ea;
      border-radius: 14px;
      transition:
        background-color $time-fast $ease-out,
        border-color $time-fast $ease-out,
        box-shadow $time-fast $ease-out;

      &::placeholder {
        color: #52636d;
      }

      &:focus {
        background-color: #ffffff;
        border-color: rgba($annonce-color-blue, 0.36);
        box-shadow: 0 0 0 3px rgba($annonce-color-blue, 0.08);
      }
    }
  }

  &__composer-actions {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: space-between;

    button {
      min-height: 34px;
      padding: 0 15px;
      font-size: 13px;
      font-weight: 800;
      border-radius: 999px;
      transition:
        background-color $time-fast $ease-out,
        opacity $time-fast $ease-out,
        transform $time-fast $ease-out;

      &:disabled {
        opacity: 0.5;
      }
    }

    button:not(.card-detail__mention-button) {
      color: #ffffff;
      background-color: $annonce-color-blue;
    }
  }

  &__mention-button {
    display: inline-flex;
    gap: 6px;
    align-items: center;
    color: $annonce-color-ink;
    background-color: #ffffff;
    border: 1px solid $annonce-color-navy;

    :deep(svg) {
      width: 14px;
      height: 14px;
      color: $annonce-color-blue;
    }
  }

  &__composer-drag-zone {
    position: absolute;
    top: 0;
    right: 0;
    left: 0;
    display: grid;
    place-items: center;
    height: 44px;
    touch-action: none;
    cursor: grab;

    &:active {
      cursor: grabbing;
    }
  }

  &__composer-handle {
    width: 40px;
    height: 5px;
    background-color: #d9d9d9;
    border-radius: 999px;
  }

  &__actions {
    position: absolute;
    right: 0;
    bottom: calc(var(--tg-safe-bottom, 0px) + 90px);
    left: 0;
    z-index: 30;
    display: flex;
    gap: 6px;
    align-items: center;
    padding: 10px 14px;
    overflow-x: auto;
    scrollbar-width: none;
    background-color: #e9eff3;
    animation: mini-rise-in 200ms $ease-out both;

    &::-webkit-scrollbar {
      display: none;
    }
  }

  &__action {
    display: inline-grid;
    grid-auto-flow: column;
    gap: 5px;
    align-items: center;
    min-width: max-content;
    min-height: 34px;
    padding: 0 15px;
    font-size: 13.5px;
    font-weight: 800;
    color: $annonce-color-ink;
    background-color: #ffffff;
    border: 1px solid $annonce-color-navy;
    border-radius: 999px;
    transition:
      background-color $time-fast $ease-out,
      border-color $time-fast $ease-out,
      box-shadow $time-fast $ease-out,
      color $time-fast $ease-out,
      transform $time-fast $ease-out;

    &:disabled {
      opacity: 0.48;
    }

    :deep(svg) {
      width: 14px;
      height: 14px;
      color: $annonce-color-blue;
    }
  }

  &__action--primary {
    background-color: #d2e2d4;
    border-color: #2e8c5a;

    :deep(svg) {
      color: #2e8c5a;
    }
  }

  &__action--active {
    color: #ffffff;
    background-color: $annonce-color-blue;
    border-color: $annonce-color-navy;

    :deep(svg) {
      color: #ffffff;
    }
  }

  &__action--icon {
    place-content: center;
    width: 44px;
    padding: 0;
  }

  h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 800;
    line-height: 24px;
    overflow-wrap: anywhere;
  }

  h3 {
    margin: 0;
    font-size: 14.5px;
    font-weight: 800;
    line-height: 20px;
  }

  :deep(.bottom-nav) {
    position: absolute;
    right: 0;
    bottom: 0;
    left: 0;
  }
}

.card-detail__composer-backdrop--dragging .card-detail__composer {
  transition: none;
}

@media (hover: hover) {
  .card-detail__nav-button:not(:disabled):hover,
  .card-detail__action:not(:disabled):hover,
  .card-detail__composer-actions button:not(:disabled):hover {
    box-shadow: 0 10px 22px rgba(#0b304b, 0.1);
    transform: translate3d(0, -1px, 0);
  }

  .card-detail__nav-button:not(:disabled):hover {
    background-color: rgba(#ffffff, 0.22);
    border-color: rgba(#ffffff, 0.4);
  }

  .card-detail__task-row:not(:disabled):hover {
    border-color: rgba($annonce-color-blue, 0.3);
    box-shadow: 0 10px 22px rgba(#0b304b, 0.08);
    transform: translate3d(0, -1px, 0);
  }

  .card-detail__task-row:not(:disabled):hover .card-detail__task-check {
    transform: scale(1.03);
  }
}

@media (prefers-reduced-motion: reduce) {
  .card-detail,
  .card-detail__header,
  .card-detail__section,
  .card-detail__assignee,
  .card-detail__task-row,
  .card-detail__comment-row,
  .card-detail__actions {
    animation: none;
  }

  .card-detail__body {
    scroll-behavior: auto;
  }

  .card-detail__nav-button,
  .card-detail__task-row,
  .card-detail__task-check,
  .card-detail__composer,
  .card-detail__composer textarea,
  .card-detail__composer-actions button,
  .card-detail__action {
    transition-duration: 1ms;
  }
}

@keyframes card-detail-fade {
  from {
    opacity: 0;
    transform: translate3d(0, 12px, 0);
  }

  to {
    opacity: 1;
    transform: translate3d(0, 0, 0);
  }
}
</style>
