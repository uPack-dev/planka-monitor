import { relativePastLabel } from '@/utils/miniDate';

const taskTabs = [
  { key: 'list', label: 'Список' },
  { key: 'calendar', label: 'Календар' },
];
const feedTabs = [
  { key: 'all', label: 'Усі' },
  { key: 'comments', label: 'Коментарі' },
  { key: 'assignments', label: 'Призначення' },
  { key: 'done', label: 'Зміни' },
];
export function useMiniAppNavigation({
  activeTab,
  clearTaskActionItem,
  currentStep,
  loadTasks,
  miniApp,
  withAppError,
}) {
  const taskView = shallowRef('list');
  const feedFilter = shallowRef('all');
  const adminStatus = shallowRef('all');
  const searchDraft = shallowRef('');
  const searchQuery = shallowRef('');
  const searchOpen = shallowRef(false);

  const topBarTitle = computed(() => {
    if (activeTab.value === 'tasks') return 'Шкала задач';
    if (activeTab.value === 'feed') return 'Стрічка подій';
    if (activeTab.value === 'admin') return 'Панель адміна';
    return 'Профіль';
  });
  const topBarEyebrow = computed(() =>
    miniApp.me?.isAdmin ? 'ANNONCE · ADMIN' : 'ANNONCE · WORKSPACE',
  );
  const topBarTabs = computed(() => {
    if (activeTab.value === 'tasks') return taskTabs;
    if (activeTab.value === 'feed') return feedTabs;
    return [];
  });
  const topBarModel = computed(() => {
    if (activeTab.value === 'tasks') return taskView.value;
    if (activeTab.value === 'feed') return feedFilter.value;
    if (activeTab.value === 'admin') return adminStatus.value;
    return '';
  });
  const searchPlaceholder = computed(() => {
    if (activeTab.value === 'admin') return 'Пошук @tg або workspace…';
    if (activeTab.value === 'feed') return 'Пошук у стрічці';
    return 'Пошук задачі, дошки або проєкту';
  });
  const adminStats = computed(() => {
    return miniApp.adminSubscribers.reduce(
      (acc, item) => {
        acc.total += 1;
        if (item.isBlocked) {
          acc.blocked += 1;
        } else if (item.isAdmin) {
          acc.admins += 1;
        } else {
          acc.users += 1;
        }
        return acc;
      },
      { total: 0, users: 0, admins: 0, blocked: 0 },
    );
  });

  watch(activeTab, async () => {
    if (activeTab.value === 'admin' && !miniApp.me?.isAdmin) {
      activeTab.value = 'tasks';
      return;
    }

    searchOpen.value = false;
    searchDraft.value = searchQuery.value;
    clearTaskActionItem();
    await loadActiveTab();
  });

  function updateTopBarModel(value) {
    if (activeTab.value === 'tasks') {
      taskView.value = value;
      loadTasks();
    } else if (activeTab.value === 'feed') {
      feedFilter.value = value;
      loadFeed();
    } else if (activeTab.value === 'admin') {
      adminStatus.value = value;
      loadAdminSubscribers();
    }
  }

  function toggleSearch() {
    if (activeTab.value === 'profile') return;
    searchDraft.value = searchQuery.value;
    searchOpen.value = !searchOpen.value;
  }

  function applySearch() {
    searchQuery.value = searchDraft.value.trim();
    if (activeTab.value === 'tasks') loadTasks();
    if (activeTab.value === 'admin') loadAdminSubscribers();
    if (activeTab.value === 'feed') loadFeed();
  }

  async function refreshActiveTab() {
    await loadActiveTab();
  }

  async function loadActiveTab() {
    if (currentStep.value !== 'app') return;
    if (activeTab.value === 'tasks') return loadTasks();
    if (activeTab.value === 'feed') return loadFeed();
    if (activeTab.value === 'profile') return loadFeed();
    if (activeTab.value === 'admin' && miniApp.me?.isAdmin) {
      return loadAdminSubscribers();
    }
  }

  async function loadFeed() {
    await withAppError('Не вдалося завантажити стрічку', () =>
      miniApp.fetchFeed({
        type: feedFilter.value,
        q: searchQuery.value,
        limit: 50,
      }),
    );
  }

  async function loadAdminSubscribers() {
    if (!miniApp.me?.isAdmin) return;
    await withAppError('Не вдалося завантажити підписників', () =>
      miniApp.fetchAdminSubscribers({
        q: searchQuery.value,
        status: adminStatus.value === 'all' ? '' : adminStatus.value,
      }),
    );
  }

  async function setSubscriberBlocked(telegramUsername, enabled) {
    await withAppError('Не вдалося оновити блокування', () =>
      miniApp.setSubscriberBlocked(telegramUsername, enabled),
    );
  }

  async function setSubscriberAdmin(telegramUsername, enabled) {
    await withAppError('Не вдалося оновити admin-доступ', () =>
      miniApp.setSubscriberAdmin(telegramUsername, enabled),
    );
  }

  function resetAdminFilters() {
    searchQuery.value = '';
    searchDraft.value = '';
    adminStatus.value = 'all';
    loadAdminSubscribers();
  }

  function feedTitle(item) {
    const actor = item.actor?.name || item.actor?.username || 'Workspace';
    const isSelf = isCurrentUserActor(item.actor);

    if (isSelf) {
      if (item.type === 'commentCard') return 'Ви прокоментували картку';
      if (item.type === 'mentionInComment') {
        return 'Ви згадали себе у коментарі';
      }
      if (item.type === 'addMemberToCard') {
        return 'Ви призначили себе на задачу';
      }
      if (item.type === 'completeTask') return 'Ви виконали задачу';
      if (item.type === 'uncompleteTask') {
        return 'Ви повернули задачу в роботу';
      }
      if (item.type === 'moveCard') return 'Ви оновили статус картки';
      return 'Ви оновили картку';
    }

    if (item.type === 'commentCard') {
      return `${actor} прокоментував картку`;
    }
    if (item.type === 'mentionInComment') {
      return `${actor} згадав вас у коментарі`;
    }
    if (item.type === 'addMemberToCard') {
      return `${actor} призначив вас на задачу`;
    }
    if (item.type === 'completeTask') {
      return `${actor} виконав вашу задачу`;
    }
    if (item.type === 'uncompleteTask') {
      return `${actor} повернув задачу в роботу`;
    }
    if (item.type === 'moveCard') {
      return `${actor} оновив статус картки`;
    }
    return `${actor} оновив картку`;
  }

  function isCurrentUserActor(actor) {
    const actorUsername = normalizeIdentity(actor?.username);
    const workspaceUsername = normalizeIdentity(miniApp.me?.workspaceUsername);
    if (actorUsername && workspaceUsername) {
      return actorUsername === workspaceUsername;
    }

    const actorName = normalizeIdentity(actor?.name);
    const workspaceDisplayName = normalizeIdentity(
      miniApp.me?.workspaceDisplayName,
    );
    if (!actorName || !workspaceDisplayName) return false;
    return actorName === workspaceDisplayName;
  }

  function normalizeIdentity(value) {
    return String(value || '')
      .trim()
      .replace(/^@+/, '')
      .toLowerCase();
  }

  return {
    adminStats,
    adminStatus,
    applySearch,
    feedFilter,
    feedTitle,
    loadAdminSubscribers,
    loadActiveTab,
    loadFeed,
    refreshActiveTab,
    relativeDate: relativePastLabel,
    resetAdminFilters,
    searchDraft,
    searchOpen,
    searchPlaceholder,
    searchQuery,
    setSubscriberAdmin,
    setSubscriberBlocked,
    taskView,
    toggleSearch,
    topBarEyebrow,
    topBarModel,
    topBarTabs,
    topBarTitle,
    updateTopBarModel,
  };
}
