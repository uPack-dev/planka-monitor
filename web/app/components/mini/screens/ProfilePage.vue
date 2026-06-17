<template>
  <section class="screen profile-screen">
    <section class="profile-hero">
      <MiniAvatar
        class="profile-hero__avatar"
        :src="avatarSource"
        :name="displayName"
        color="#ffc261"
        size="xxl"
      />
      <h1>{{ displayName }}</h1>
      <p>@{{ telegramUsername }} · Workspace: {{ workspaceLabel }}</p>
    </section>

    <section class="profile-stats" aria-label="Статистика профілю">
      <article
        v-for="(item, itemIndex) in profileStats"
        :key="item.key"
        class="profile-stat"
        :class="`profile-stat--${item.tone}`"
        :style="{ '--motion-delay': `${itemIndex * 20}ms` }"
      >
        <strong>{{ item.value }}</strong>
        <p>
          {{ item.label }}
          <span v-if="item.accent">{{ item.accent }}</span>
        </p>
      </article>
    </section>

    <component
      :is="workspaceUrl ? 'a' : 'section'"
      class="workspace-card"
      :href="workspaceUrl || undefined"
      :target="workspaceUrl ? '_blank' : undefined"
      :rel="workspaceUrl ? 'noopener noreferrer' : undefined"
    >
      <span class="workspace-card__icon">
        <CIcon name="workspace-mini" />
      </span>
      <span class="workspace-card__copy">
        <strong>Workspace</strong>
        <small>{{ workspaceHandle }}</small>
      </span>
    </component>

    <p v-if="errorCode" class="screen-error">
      {{ preferenceError }}
    </p>
  </section>
</template>

<script setup>
const props = defineProps({
  avatarSource: {
    type: String,
    required: true,
  },
  displayName: {
    type: String,
    required: true,
  },
  telegramUsername: {
    type: String,
    required: true,
  },
  workspaceLabel: {
    type: String,
    required: true,
  },
  workspaceUrl: {
    type: String,
    default: '',
  },
  taskStats: {
    type: Object,
    default: () => ({
      overdue: 0,
      today: 0,
      planned: 0,
    }),
  },
  completedTodayCount: {
    type: Number,
    default: 0,
  },
  completedMonthCount: {
    type: Number,
    default: 0,
  },
  dueThisMonthCount: {
    type: Number,
    default: 0,
  },
  unreadFeedCount: {
    type: Number,
    default: 0,
  },
  notificationsEnabled: {
    type: Boolean,
    required: true,
  },
  enabledPreferenceCount: {
    type: Number,
    required: true,
  },
  isAdmin: {
    type: Boolean,
    required: true,
  },
  masterNotifications: {
    type: Boolean,
    required: true,
  },
  preferenceCards: {
    type: Array,
    required: true,
  },
  preferences: {
    type: Object,
    required: true,
  },
  errorCode: {
    type: String,
    default: '',
  },
  preferenceError: {
    type: String,
    required: true,
  },
});

defineEmits(['toggle-master', 'update-preference']);

const workspaceHandle = computed(() =>
  props.workspaceLabel === 'не привʼязано'
    ? props.workspaceLabel
    : formatWorkspaceHandle(props.workspaceLabel, props.workspaceUrl),
);

const profileStats = computed(() => [
  {
    key: 'feed',
    value: formatCount(props.unreadFeedCount),
    label: 'нових повідомлень в стрічці подій',
    tone: 'feed',
  },
  {
    key: 'overdue',
    value: formatCount(props.taskStats.overdue),
    label: 'задач мають статус прострочених',
    tone: 'danger',
  },
  {
    key: 'completed-today',
    value: formatCount(props.completedTodayCount),
    label: 'задач виконано',
    accent: 'за сьогодні',
    tone: 'plain',
  },
  {
    key: 'due-today',
    value: formatCount(props.taskStats.today),
    label: 'задачі мають бути виконані',
    accent: 'сьогодні',
    tone: 'today',
  },
  {
    key: 'completed-month',
    value: formatCount(props.completedMonthCount),
    label: 'задач виконано',
    accent: 'в цьому місяці',
    tone: 'plain',
  },
  {
    key: 'due-month',
    value: formatCount(props.dueThisMonthCount),
    label: 'задач мають бути виконані',
    accent: 'до кінця місяця',
    tone: 'plain',
  },
]);

function formatCount(value) {
  const count = Number(value);
  return Number.isFinite(count) && count > 0 ? Math.round(count) : 0;
}

function formatWorkspaceHandle(username, workspaceUrl) {
  const cleanUsername = String(username || '')
    .trim()
    .replace(/^@+/, '');
  if (!cleanUsername) return '';

  const host = workspaceHost(workspaceUrl);
  return host ? `${cleanUsername}@${host}` : `@${cleanUsername}`;
}

function workspaceHost(workspaceUrl) {
  try {
    return new URL(workspaceUrl).host;
  } catch {
    return '';
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

.screen-error {
  font-size: 13px;
  line-height: 1.35;
  color: $annonce-color-error;
  text-align: center;
}

.profile-screen {
  gap: 14px;
  margin-top: -16px;
}

.profile-hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-height: 248px;
  padding: 14px 16px 22px;
  margin-right: -16px;
  margin-left: -16px;
  text-align: center;
  background-color: $annonce-color-navy;
  animation: mini-rise-in 240ms $ease-out both;

  &__avatar {
    --mini-avatar-size: 150px;
    --mini-avatar-font-size: 48px;

    margin-bottom: 19px;
    font-weight: 700;
    color: $color-white;
    box-shadow: 0 18px 48px rgba(#031d31, 0.24);
    animation: mini-soft-pop 230ms $ease-out both;
  }

  h1 {
    max-width: 100%;
    margin-top: auto;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: $annonce-font-brand;
    font-size: 17px;
    font-weight: 400;
    line-height: 21px;
    color: #bdd0dc;
    white-space: nowrap;
  }

  p {
    max-width: 100%;
    margin-top: 5px;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
    font-weight: 500;
    line-height: 17px;
    color: rgba($annonce-color-muted, 0.92);
    white-space: nowrap;
  }
}

.profile-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 10px;
}

.profile-stat {
  display: grid;
  gap: 4px;
  align-content: start;
  min-width: 0;
  min-height: 85px;
  padding: 10px;
  background-color: $color-white;
  border: 1.5px solid #c9dde8;
  border-radius: 14px;
  transition:
    border-color $time-fast $ease-out,
    box-shadow $time-fast $ease-out,
    transform $time-fast $ease-out;
  animation: mini-rise-in 220ms $ease-out both;
  animation-delay: var(--motion-delay, 0ms);

  strong {
    font-family: $annonce-font-brand;
    font-size: 22px;
    font-weight: 400;
    line-height: 24px;
    color: $annonce-color-blue;
    overflow-wrap: anywhere;
  }

  p {
    font-size: 12px;
    font-weight: 500;
    line-height: 17px;
    color: $annonce-color-ink-muted;
    overflow-wrap: anywhere;

    span {
      font-weight: 700;
    }
  }

  &--feed {
    background-color: #fff5db;
    border-color: $annonce-color-yellow-action;
  }

  &--danger {
    background-color: #ffe7ed;
    border-color: #f1b6c4;
  }

  &--today {
    background-color: #e7f5ff;
    border-color: #5995c1;
  }
}

.workspace-card {
  display: grid;
  grid-template-columns: 32px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
  width: 100%;
  min-height: 56px;
  padding: 12px;
  margin-top: 37px;
  color: $color-white;
  text-decoration: none;
  background-color: $annonce-color-blue;
  border-radius: 14px;
  box-shadow: 0 12px 28px rgba(#0b304b, 0.12);
  transition:
    background-color $time-fast $ease-out,
    box-shadow $time-fast $ease-out,
    transform $time-fast $ease-out;
  animation: mini-rise-in 220ms $ease-out both;
  animation-delay: 20ms;

  &:focus-visible {
    outline: 2px solid $annonce-color-yellow;
    outline-offset: 3px;
  }

  &__icon {
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    color: $annonce-color-blue;
    background-color: #e1edf5;
    border-radius: 10px;
    transition: transform $time-fast $ease-out;

    :deep(svg) {
      width: 18px;
      height: 18px;
    }
  }

  &__copy {
    display: grid;
    gap: 2px;
    min-width: 0;
  }

  strong,
  small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    font-size: 15px;
    font-weight: 700;
    line-height: 20px;
  }

  small {
    font-size: 13px;
    font-weight: 500;
    line-height: 18px;
  }
}

@media (hover: hover) {
  .profile-stat:hover {
    border-color: rgba($annonce-color-blue, 0.32);
    box-shadow: 0 12px 24px rgba(#0b304b, 0.09);
    transform: translate3d(0, -1px, 0);
  }

  .workspace-card:hover {
    background-color: $annonce-color-navy;
    box-shadow: 0 16px 34px rgba(#0b304b, 0.16);
    transform: translate3d(0, -1px, 0);
  }

  .workspace-card:hover .workspace-card__icon {
    transform: rotate(-4deg) scale(1.02);
  }
}

@media (prefers-reduced-motion: reduce) {
  .profile-hero,
  .profile-hero__avatar,
  .profile-stat,
  .workspace-card {
    animation: none;
  }

  .profile-stat,
  .workspace-card,
  .workspace-card__icon {
    transition-duration: 1ms;
  }
}

@media (width <= 360px) {
  .profile-stat {
    min-height: 90px;
  }

  .profile-stat p {
    font-size: 11.5px;
  }
}
</style>
