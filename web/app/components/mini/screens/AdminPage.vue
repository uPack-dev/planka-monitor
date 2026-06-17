<template>
  <section class="screen admin-screen">
    <MiniEmptyState
      v-if="!isAdmin"
      icon="settings"
      title="Тільки для адмінів"
      text="Керування підписниками доступне monitor-admin користувачам."
    />

    <template v-else>
      <div class="admin-summary" aria-label="Статистика підписників">
        <span class="admin-summary__chip admin-summary__chip--users">
          Користувачі {{ normalizedStats.users }}
        </span>
        <span class="admin-summary__chip admin-summary__chip--admins">
          Адмін {{ normalizedStats.admins }}
        </span>
        <span class="admin-summary__chip admin-summary__chip--blocked">
          Заблоковано {{ normalizedStats.blocked }}
        </span>
      </div>

      <div v-if="isLoading && !items.length" class="loader">
        Оновлюємо підписників...
      </div>

      <MiniEmptyState
        v-else-if="!items.length"
        icon="user"
        title="Нікого не знайдено"
        text="Змініть пошук або фільтр статусу."
        action-label="Скинути"
        @action="$emit('reset')"
      />

      <div v-else class="subscribers-list">
        <MiniSubscriberRow
          v-for="(item, itemIndex) in items"
          :key="item.telegramUsername"
          :item="item"
          :style="{ '--motion-delay': `${itemIndex * 18}ms` }"
          @block="$emit('block', item.telegramUsername, $event)"
          @admin="$emit('admin', item.telegramUsername, $event)"
          @menu="openSubscriberMenu"
        />
      </div>

      <MiniSubscriberActionSheet
        :item="actionItem"
        @close="closeSubscriberMenu"
        @admin="adminActionSubscriber"
        @block="blockActionSubscriber"
      />
    </template>
  </section>
</template>

<script setup>
const props = defineProps({
  isAdmin: {
    type: Boolean,
    required: true,
  },
  isLoading: {
    type: Boolean,
    required: true,
  },
  items: {
    type: Array,
    required: true,
  },
  stats: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(['reset', 'block', 'admin']);

const actionItem = shallowRef(null);
const normalizedStats = computed(() => {
  const blocked = toCount(props.stats.blocked);
  const admins = toCount(props.stats.admins);
  const total = toCount(props.stats.total);
  const users =
    props.stats.users === undefined
      ? Math.max(0, total - admins - blocked)
      : toCount(props.stats.users);

  return { users, admins, blocked };
});

function openSubscriberMenu(item) {
  actionItem.value = item;
}

function closeSubscriberMenu() {
  actionItem.value = null;
}

function adminActionSubscriber(item) {
  if (!item?.telegramUsername) return;
  emit('admin', item.telegramUsername, !item.isAdmin);
  closeSubscriberMenu();
}

function blockActionSubscriber(item) {
  if (!item?.telegramUsername) return;
  emit('block', item.telegramUsername, !item.isBlocked);
  closeSubscriberMenu();
}

function toCount(value) {
  const count = Number(value);
  return Number.isFinite(count) && count > 0 ? Math.round(count) : 0;
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

.admin-summary {
  display: flex;
  gap: 10px;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }

  &__chip {
    display: grid;
    flex: 0 0 auto;
    place-items: center;
    height: 24px;
    padding: 0 16px;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
    font-weight: 700;
    line-height: 16px;
    white-space: nowrap;
    border: 1px solid;
    border-radius: 999px;
    transition:
      box-shadow $time-fast $ease-out,
      transform $time-fast $ease-out;
    animation: mini-rise-in 200ms $ease-out both;
  }

  &__chip--users {
    min-width: 125px;
    color: #2e8c5a;
    background-color: rgba(#2e8c5a, 0.1);
    border-color: rgba(#2e8c5a, 0.4);
    animation-delay: 18ms;
  }

  &__chip--admins {
    min-width: 100px;
    color: $annonce-color-blue;
    background-color: rgba($annonce-color-blue, 0.1);
    border-color: rgba($annonce-color-blue, 0.4);
    animation-delay: 35ms;
  }

  &__chip--blocked {
    min-width: 125px;
    color: #a3263e;
    background-color: #ffe7ed;
    border-color: #f1b6c4;
    animation-delay: 20ms;
  }
}

.subscribers-list {
  display: grid;
  gap: 10px;
}

@media (hover: hover) {
  .admin-summary__chip:hover {
    box-shadow: 0 8px 18px rgba(#0b304b, 0.08);
    transform: translate3d(0, -1px, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .admin-summary__chip {
    transition-duration: 1ms;
    animation: none;
  }
}
</style>
