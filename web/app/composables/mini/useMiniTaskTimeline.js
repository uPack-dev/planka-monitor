import {
  addDays,
  dateKey,
  dayLabel,
  endOfDay,
  endOfMonth,
  formatClockTime,
  getCalendarYearRangeStart,
  mondayIndex,
  parseDateKey,
  parseValidDate,
  startOfDay,
  startOfMonth,
} from '@/utils/miniDate';
import {
  MIN_DESIGN_TASK_COUNT,
  isMockTasksEnabled,
  mockTaskToCard,
} from '@/utils/mockTasks.mjs';

const TASK_LIST_HISTORY_START = '2000-01-01T00:00:00.000Z';
const calendarDayNames = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
const calendarMonthOptions = Array.from({ length: 12 }, (_, value) => ({
  value,
  label: capitalize(
    new Date(2026, value, 1).toLocaleDateString('uk-UA', { month: 'long' }),
  ),
}));

export function useMiniTaskTimeline({
  activeTab,
  displayName,
  miniApp,
  route,
  runtimeConfig,
  searchQuery,
  taskView,
  withAppError,
}) {
  const selectedDate = shallowRef(dateKey(new Date()));
  const calendarViewDate = shallowRef(selectedDate.value);
  const calendarPickerMode = shallowRef('');
  const calendarYearPickerStart = shallowRef(
    getCalendarYearRangeStart(new Date()),
  );
  const cardComposerRequestKey = shallowRef(0);
  const taskActionItem = shallowRef(null);
  const feedActionItem = shallowRef(null);

  const currentWeekday = computed(() =>
    new Date().toLocaleDateString('uk-UA', { weekday: 'long' }),
  );
  const currentDateLabel = computed(() =>
    new Date().toLocaleDateString('uk-UA', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
    }),
  );
  const greetingLabel = computed(() => {
    const name = (displayName.value || 'друже').split(/\s|_/)[0];
    const hour = new Date().getHours();
    const greeting = hour < 12 ? 'Добрий ранок' : 'Добрий день';

    return `${greeting}, ${name}!`;
  });
  const mockTasksMode = computed(() =>
    String(route.query.mockTasks ?? runtimeConfig.public.mockTasks ?? '')
      .trim()
      .toLowerCase(),
  );
  const shouldUseDesignMockTasks = computed(() =>
    isMockTasksEnabled(mockTasksMode.value),
  );
  const minDesignTaskCount = computed(() => {
    const value = Number(
      route.query.mockTaskCount || runtimeConfig.public.mockTasksMin,
    );
    return Number.isFinite(value) && value > 0
      ? Math.round(value)
      : MIN_DESIGN_TASK_COUNT;
  });
  const selectedDateObject = computed(() => parseDateKey(selectedDate.value));
  const calendarViewDateObject = computed(() =>
    parseDateKey(calendarViewDate.value),
  );
  const calendarMonthLabel = computed(() =>
    capitalize(
      calendarViewDateObject.value.toLocaleDateString('uk-UA', {
        month: 'long',
      }),
    ),
  );
  const calendarYearLabel = computed(() =>
    calendarViewDateObject.value.toLocaleDateString('uk-UA', {
      year: 'numeric',
    }),
  );
  const calendarYearOptions = computed(() =>
    Array.from(
      { length: 10 },
      (_, index) => calendarYearPickerStart.value + index,
    ),
  );
  const calendarYearRangeLabel = computed(
    () =>
      `${calendarYearPickerStart.value}-${calendarYearPickerStart.value + 9}`,
  );
  const calendarTaskCountByDate = computed(() => {
    const counts = new Map();
    for (const item of miniApp.timelineItems) {
      const due = parseItemDueDate(item);
      if (!due) continue;
      const key = dateKey(due);
      counts.set(key, (counts.get(key) || 0) + 1);
    }
    return counts;
  });
  const monthCalendarDays = computed(() => {
    const monthStart = startOfMonth(calendarViewDateObject.value);
    const monthEnd = endOfMonth(calendarViewDateObject.value);
    const gridStart = addDays(monthStart, -mondayIndex(monthStart));
    const gridEnd = addDays(monthEnd, 6 - mondayIndex(monthEnd));
    const length =
      Math.round((gridEnd.getTime() - gridStart.getTime()) / 86400000) + 1;
    const today = dateKey(new Date());

    return Array.from({ length }, (_, index) => {
      const date = addDays(gridStart, index);
      const key = dateKey(date);
      const isCurrentMonth =
        date.getFullYear() === monthStart.getFullYear() &&
        date.getMonth() === monthStart.getMonth();

      return {
        key,
        day: date.getDate(),
        isCurrentMonth,
        isToday: isCurrentMonth && key === today,
        taskCount: calendarTaskCountByDate.value.get(key) || 0,
      };
    });
  });
  const visibleTaskItems = computed(() => {
    if (taskView.value !== 'calendar') return miniApp.timelineItems;
    return miniApp.timelineItems.filter((item) => {
      const due = parseItemDueDate(item);
      return due && dateKey(due) === selectedDate.value;
    });
  });
  const selectedDateLabel = computed(() =>
    dayLabel(selectedDateObject.value, {
      capitalized: true,
      includeYesterday: false,
    }),
  );
  const calendarRange = computed(() => {
    const monthStart = startOfMonth(calendarViewDateObject.value);
    const monthEnd = endOfMonth(calendarViewDateObject.value);
    return {
      from: addDays(monthStart, -mondayIndex(monthStart)),
      to: endOfDay(addDays(monthEnd, 6 - mondayIndex(monthEnd))),
    };
  });
  const taskRange = computed(() => {
    if (taskView.value === 'calendar') {
      return calendarRange.value;
    }
    const today = startOfDay(new Date());
    return {
      from: new Date(TASK_LIST_HISTORY_START),
      to: endOfDay(addDays(today, 60)),
    };
  });
  const taskGroups = computed(() => {
    const groups = new Map();
    for (const item of visibleTaskItems.value) {
      const due = parseItemDueDate(item);
      const key = due ? dateKey(due) : 'undated';
      if (!groups.has(key)) {
        groups.set(key, {
          key,
          label: due ? groupDateLabel(due) : 'Без дати',
          tone: due ? timelineTone(due) : 'undated',
          items: [],
        });
      }
      groups.get(key).items.push(item);
    }
    return Array.from(groups.values());
  });
  const taskStats = computed(() => {
    const today = dateKey(new Date());
    return miniApp.timelineItems.reduce(
      (acc, item) => {
        const due = parseItemDueDate(item);
        if (!due) {
          acc.planned += 1;
          return acc;
        }

        const dueKey = dateKey(due);
        if (dueKey < today) acc.overdue += 1;
        else if (dueKey === today) acc.today += 1;
        else acc.planned += 1;
        return acc;
      },
      {
        overdue: 0,
        today: 0,
        planned: 0,
      },
    );
  });
  const dueThisMonthCount = computed(() => {
    const today = startOfDay(new Date());
    const monthEnd = endOfMonth(today);

    return miniApp.timelineItems.reduce((count, item) => {
      const due = parseItemDueDate(item);
      if (!due || due < today || due > monthEnd) return count;
      return count + 1;
    }, 0);
  });
  const activeCardWorkspaceUrl = computed(() => {
    if (!miniApp.activeCard?.id) return '';
    if (miniApp.activeCard.cardWorkspaceUrl) {
      return miniApp.activeCard.cardWorkspaceUrl;
    }
    const item = miniApp.timelineItems.find(
      (task) => task.cardId === miniApp.activeCard.id,
    );
    return item?.cardWorkspaceUrl || item?.workspaceUrl || '';
  });

  async function loadTasks() {
    const { from, to } = taskRange.value;
    const today = startOfDay(new Date());
    await withAppError('Не вдалося завантажити задачі', () =>
      miniApp.fetchTasks({
        from: from.toISOString(),
        to: to.toISOString(),
        statsFrom: today.toISOString(),
        statsTo: endOfDay(today).toISOString(),
        q: searchQuery.value,
        view: taskView.value,
        mockTasks: shouldUseDesignMockTasks.value && !searchQuery.value,
        minMockTasks: minDesignTaskCount.value,
        mockSelectedDate: selectedDate.value,
      }),
    );
  }

  function selectDate(key) {
    closeCalendarPicker();
    if (!isDateKey(key)) {
      return;
    }

    const didChangeViewDate = setCalendarViewDate(key);
    if (selectedDate.value === key) {
      if (didChangeViewDate) loadTasks();
      return;
    }

    selectedDate.value = key;
    loadTasks();
  }

  function toggleCalendarPicker(mode) {
    calendarPickerMode.value = calendarPickerMode.value === mode ? '' : mode;

    if (calendarPickerMode.value === 'year') {
      calendarYearPickerStart.value = getCalendarYearRangeStart(
        calendarViewDateObject.value,
      );
    }
  }

  function closeCalendarPicker() {
    calendarPickerMode.value = '';
  }

  function setCalendarViewDate(value) {
    const nextDateKey = toDateKey(value);
    if (!isDateKey(nextDateKey) || calendarViewDate.value === nextDateKey) {
      return false;
    }

    calendarViewDate.value = nextDateKey;
    return true;
  }

  function selectCalendarMonth(month) {
    selectCalendarPeriodPart({
      month,
      year: calendarViewDateObject.value.getFullYear(),
    });
  }

  function selectCalendarYear(year) {
    selectCalendarPeriodPart({
      month: calendarViewDateObject.value.getMonth(),
      year,
    });
  }

  function selectCalendarPeriodPart({ month, year }) {
    closeCalendarPicker();
    const current = calendarViewDateObject.value;
    const nextMonthEnd = endOfMonth(new Date(year, month, 1));
    const nextDay = Math.min(current.getDate(), nextMonthEnd.getDate());

    if (setCalendarViewDate(new Date(year, month, nextDay))) {
      loadTasks();
    }
  }

  function changeCalendarYearPage(direction) {
    calendarYearPickerStart.value += direction * 10;
  }

  function changeCalendarMonth(direction) {
    closeCalendarPicker();
    const current = calendarViewDateObject.value;
    const nextMonth = new Date(
      current.getFullYear(),
      current.getMonth() + direction,
      1,
    );
    const nextMonthEnd = endOfMonth(nextMonth);
    const nextDay = Math.min(current.getDate(), nextMonthEnd.getDate());

    if (
      setCalendarViewDate(
        new Date(nextMonth.getFullYear(), nextMonth.getMonth(), nextDay),
      )
    ) {
      loadTasks();
    }
  }

  async function openCard(item) {
    if (item.isMock) {
      miniApp.activeCard = mockTaskToCard(item);
      return;
    }

    await withAppError('Не вдалося відкрити картку', () =>
      miniApp.fetchCard(item.cardId),
    );
  }

  async function openCardForComment(item) {
    await openCard(item);
    if (miniApp.activeCard?.id === item?.cardId) {
      cardComposerRequestKey.value += 1;
    }
  }

  function openTaskMenu(item) {
    taskActionItem.value = item;
  }

  function openFeedMenu(item) {
    feedActionItem.value = item;
  }

  function clearTaskActionItem() {
    taskActionItem.value = null;
    feedActionItem.value = null;
  }

  async function completeTaskActionItem(item) {
    taskActionItem.value = null;
    await completeTimelineItem(item);
  }

  async function commentTaskActionItem(item) {
    taskActionItem.value = null;
    await openCardForComment(item);
  }

  async function muteTaskActionItem(item) {
    await setTimelineItemMuted(item, !item?.isMuted);
  }

  async function muteFeedActionItem(item) {
    await setFeedItemMuted(item, !item?.isMuted);
  }

  function navigateFromCardSheet(tab) {
    miniApp.activeCard = null;
    activeTab.value = tab;
  }

  async function openFeedItem(item) {
    if (!item.isRead) {
      await miniApp.markFeedRead(item.id).catch(() => {});
    }
    if (item.cardId) {
      await openCard({ cardId: item.cardId });
    }
  }

  async function completeCard(cardId) {
    if (isMockId(cardId)) {
      completeMockCard(cardId);
      return;
    }

    if (
      await withAppError('Не вдалося позначити картку виконаною', () =>
        miniApp.completeCard(cardId),
      )
    ) {
      miniApp.activeCard = null;
    }
  }

  async function completeTimelineItem(item) {
    if (item?.taskId) {
      await completeTask(item.taskId);
      return;
    }
    if (item?.cardId) {
      await completeCard(item.cardId);
    }
  }

  async function completeTask(taskId) {
    if (isMockId(taskId)) {
      completeMockTask(taskId);
      return;
    }

    await withAppError('Не вдалося закрити задачу', () =>
      miniApp.completeTask(taskId),
    );
  }

  async function addComment(text) {
    if (!miniApp.activeCard?.id) return;
    if (miniApp.activeCard?.isMock) {
      miniApp.activeCard = null;
      return;
    }

    await withAppError('Не вдалося додати коментар', () =>
      miniApp.addComment(miniApp.activeCard.id, text),
    );
  }

  async function setTimelineItemMuted(item, isMuted) {
    const cardId = item?.cardId;
    if (!cardId) return;
    const previousMuted = Boolean(item?.isMuted);
    if (previousMuted === isMuted) return;

    if (isMockId(cardId)) {
      setMockCardMuted(cardId, isMuted);
      return;
    }

    updateTimelineItemsByCard(cardId, { isMuted });
    updateFeedItemsByCard(cardId, { isMuted });
    syncTaskActionPatch(cardId, { isMuted, isMutePending: true });
    syncFeedActionPatch(cardId, { isMuted });

    const didUpdate = await withAppError(
      isMuted
        ? 'Не вдалося заглушити задачу'
        : 'Не вдалося увімкнути сповіщення',
      () => (isMuted ? miniApp.muteCard(cardId) : miniApp.unmuteCard(cardId)),
    );
    if (didUpdate) {
      syncTaskActionPatch(cardId, { isMuted, isMutePending: false });
      syncFeedActionPatch(cardId, { isMuted, isMutePending: false });
      return;
    }

    updateTimelineItemsByCard(cardId, { isMuted: previousMuted });
    updateFeedItemsByCard(cardId, { isMuted: previousMuted });
    syncTaskActionPatch(cardId, {
      isMuted: previousMuted,
      isMutePending: false,
    });
    syncFeedActionPatch(cardId, {
      isMuted: previousMuted,
      isMutePending: false,
    });
  }

  async function setFeedItemMuted(item, isMuted) {
    const cardId = item?.cardId;
    if (!cardId) return;
    const previousMuted = Boolean(item?.isMuted);
    if (previousMuted === isMuted) return;

    updateFeedItemsByCard(cardId, { isMuted });
    updateTimelineItemsByCard(cardId, { isMuted });
    syncFeedActionPatch(cardId, { isMuted, isMutePending: true });
    syncTaskActionPatch(cardId, { isMuted });

    const didUpdate = await withAppError(
      isMuted
        ? 'Не вдалося заглушити сповіщення'
        : 'Не вдалося увімкнути сповіщення',
      () => (isMuted ? miniApp.muteCard(cardId) : miniApp.unmuteCard(cardId)),
    );
    if (didUpdate) {
      syncFeedActionPatch(cardId, { isMuted, isMutePending: false });
      syncTaskActionPatch(cardId, { isMuted, isMutePending: false });
      return;
    }

    updateFeedItemsByCard(cardId, { isMuted: previousMuted });
    updateTimelineItemsByCard(cardId, { isMuted: previousMuted });
    syncFeedActionPatch(cardId, {
      isMuted: previousMuted,
      isMutePending: false,
    });
    syncTaskActionPatch(cardId, {
      isMuted: previousMuted,
      isMutePending: false,
    });
  }

  function syncTaskActionMute(cardId, isMuted) {
    syncTaskActionPatch(cardId, { isMuted });
  }

  function syncTaskActionPatch(cardId, patch) {
    if (taskActionItem.value?.cardId !== cardId) return;

    taskActionItem.value = {
      ...taskActionItem.value,
      ...patch,
    };
  }

  function syncFeedActionPatch(cardId, patch) {
    if (feedActionItem.value?.cardId !== cardId) return;

    feedActionItem.value = {
      ...feedActionItem.value,
      ...patch,
    };
  }

  function completeMockCard(cardId) {
    removeTimelineItems((item) => item.cardId === cardId);
    miniApp.activeCard = null;
  }

  function completeMockTask(taskId) {
    removeTimelineItems((item) => item.taskId === taskId);
    updateActiveCardTask(taskId, { isCompleted: true });
  }

  function setMockCardMuted(cardId, isMuted) {
    updateTimelineItemsByCard(cardId, { isMuted });
    updateFeedItemsByCard(cardId, { isMuted });
    syncTaskActionMute(cardId, isMuted);
    syncFeedActionPatch(cardId, { isMuted });
  }

  function removeTimelineItems(predicate) {
    miniApp.timelineItems = miniApp.timelineItems.filter(
      (item) => !predicate(item),
    );
  }

  function updateTimelineItemsByCard(cardId, patch) {
    miniApp.timelineItems = miniApp.timelineItems.map((item) =>
      item.cardId === cardId ? { ...item, ...patch } : item,
    );
  }

  function updateFeedItemsByCard(cardId, patch) {
    miniApp.feedItems = (miniApp.feedItems || []).map((item) =>
      item.cardId === cardId ? { ...item, ...patch } : item,
    );
  }

  function updateActiveCardTask(taskId, patch) {
    if (!miniApp.activeCard?.tasks) return;
    miniApp.activeCard = {
      ...miniApp.activeCard,
      tasks: miniApp.activeCard.tasks.map((task) =>
        task.id === taskId ? { ...task, ...patch } : task,
      ),
    };
  }

  return {
    activeCardWorkspaceUrl,
    addComment,
    calendarDayNames,
    calendarMonthLabel,
    calendarMonthOptions,
    calendarPickerMode,
    calendarYearLabel,
    calendarYearOptions,
    calendarYearRangeLabel,
    calendarViewDateObject,
    cardComposerRequestKey,
    changeCalendarMonth,
    changeCalendarYearPage,
    closeCalendarPicker,
    clearTaskActionItem,
    commentTaskActionItem,
    completeCard,
    completeTask,
    completeTaskActionItem,
    completeTimelineItem,
    currentDateLabel,
    currentWeekday,
    dueThisMonthCount,
    feedActionItem,
    greetingLabel,
    loadTasks,
    monthCalendarDays,
    muteTaskActionItem,
    muteFeedActionItem,
    navigateFromCardSheet,
    openCard,
    openFeedMenu,
    openFeedItem,
    openTaskMenu,
    selectCalendarMonth,
    selectCalendarYear,
    selectDate,
    selectedDate,
    selectedDateLabel,
    selectedDateObject,
    taskActionItem,
    taskGroups,
    taskStats,
    toggleCalendarPicker,
    visibleTaskItems,
  };
}

function isDateKey(value) {
  return /^\d{4}-\d{2}-\d{2}$/.test(value);
}

function toDateKey(value) {
  return value instanceof Date ? dateKey(value) : value;
}

function isMockId(value) {
  return String(value || '').startsWith('mock');
}

function capitalize(value) {
  return value ? value.slice(0, 1).toUpperCase() + value.slice(1) : '';
}

function groupDateLabel(date) {
  const label = dayLabel(date, {
    capitalized: true,
    includeYesterday: false,
  });
  const time = formatClockTime(date);

  return `${label} - ${time}`;
}

function parseItemDueDate(item) {
  return parseValidDate(item?.dueAt);
}

function timelineTone(date) {
  const key = dateKey(date);
  const today = dateKey(new Date());
  if (key < today) return 'overdue';
  if (key === today) return 'today';
  return 'future';
}
