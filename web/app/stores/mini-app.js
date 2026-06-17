import { defineStore } from 'pinia';
import { ensureMockTaskCoverage } from '@/utils/mockTasks.mjs';

export const DEFAULT_PREFERENCES = {
  assignments: true,
  comments: true,
  changes: false,
  done: true,
};

export const useMiniAppStore = defineStore('mini-app', () => {
  const me = shallowRef(null);
  const timelineItems = shallowRef([]);
  const timelineStats = shallowRef({ completed: 0 });
  const feedItems = shallowRef([]);
  const adminSubscribers = shallowRef([]);
  const activeCard = shallowRef(null);
  const isLoading = shallowRef(false);
  const errorCode = shallowRef('');
  const errorStatus = shallowRef(0);

  const preferences = computed(
    () => me.value?.preferences || DEFAULT_PREFERENCES,
  );

  const hasWorkspace = computed(() =>
    Boolean(me.value?.workspaceUsername && me.value?.workspaceExists),
  );

  async function request(url, options = {}) {
    errorCode.value = '';
    errorStatus.value = 0;
    isLoading.value = true;

    try {
      const response = await useTelegramRequest(url, options);
      return response._data;
    } catch (error) {
      const data = error?.data || error?.response?._data;
      const status =
        data?.status || error?.status || error?.response?.status || 0;

      errorStatus.value = status;
      errorCode.value =
        data?.error || (status ? `http_${status}` : 'request_failed');
      throw error;
    } finally {
      isLoading.value = false;
    }
  }

  async function fetchMe() {
    me.value = await request('/me');
    return me.value;
  }

  async function bindWorkspace(workspaceUsername) {
    me.value = await request('/me/bind', {
      method: 'POST',
      body: { workspaceUsername },
    });
    return me.value;
  }

  async function savePreferences(nextPreferences) {
    me.value = await request('/me/preferences', {
      method: 'PATCH',
      body: nextPreferences,
    });
    return me.value;
  }

  async function setMasterNotifications(enabled) {
    me.value = await request('/me/notifications', {
      method: 'PATCH',
      body: { enabled },
    });
    return me.value;
  }

  async function setOnboardingCompleted(completed) {
    me.value = await request('/me/onboarding', {
      method: 'PATCH',
      body: { completed },
    });
    return me.value;
  }

  async function fetchTasks(query = {}) {
    const { mockTasks, minMockTasks, mockNow, mockSelectedDate, ...apiQuery } =
      query;
    const data = await request('/tasks', { query: apiQuery });
    const items = data?.items || [];
    timelineStats.value = {
      completed: normalizeCount(data?.stats?.completed),
    };

    timelineItems.value = mockTasks
      ? ensureMockTaskCoverage(items, {
          from: apiQuery.from,
          to: apiQuery.to,
          minCount: minMockTasks,
          now: mockNow,
          selectedDate: mockSelectedDate,
        })
      : items;
    return timelineItems.value;
  }

  async function fetchCard(cardId) {
    activeCard.value = await request(`/cards/${cardId}`);
    return activeCard.value;
  }

  async function completeCard(cardId) {
    await request(`/cards/${cardId}/complete`, { method: 'POST' });
    const completedItem =
      timelineItems.value.find(isCardTimelineItem(cardId)) ||
      (activeCard.value?.id === cardId ? activeCard.value : null);
    removeTimelineItems(isCardTimelineItem(cardId));
    incrementCompleted(completedItem);
    updateActiveCard(cardId, { isDueCompleted: true });
  }

  async function completeTask(taskId) {
    await request(`/tasks/${taskId}/complete`, { method: 'POST' });
    const activeCardHasTask = activeCard.value?.tasks?.some(
      (task) => task.id === taskId,
    );
    const completedItem =
      timelineItems.value.find((item) => item.taskId === taskId) ||
      (activeCardHasTask ? activeCard.value : null);
    removeTimelineItems((item) => item.taskId === taskId);
    incrementCompleted(completedItem);
    updateActiveCardTask(taskId, { isCompleted: true });
  }

  async function addComment(cardId, text) {
    await request(`/cards/${cardId}/comments`, {
      method: 'POST',
      body: { text },
    });
    return fetchCard(cardId);
  }

  async function muteCard(cardId) {
    await request(`/cards/${cardId}/mute`, { method: 'POST' });
    setCardMuted(cardId, true);
  }

  async function unmuteCard(cardId) {
    await request(`/cards/${cardId}/mute`, { method: 'DELETE' });
    setCardMuted(cardId, false);
  }

  function setCardMuted(cardId, isMuted) {
    updateTimelineItemsByCard(cardId, { isMuted });
    updateFeedItemsByCard(cardId, { isMuted });
    updateActiveCard(cardId, { isMuted });
  }

  async function fetchFeed(query = {}) {
    const data = await request('/feed', { query });
    feedItems.value = data?.items || [];
    return feedItems.value;
  }

  async function markFeedRead(id) {
    await request(`/feed/${id}/read`, { method: 'PATCH' });
    updateFeedItem(id, { isRead: true });
  }

  async function fetchAdminSubscribers(query = {}) {
    const data = await request('/admin/subscribers', { query });
    adminSubscribers.value = data?.items || [];
    return adminSubscribers.value;
  }

  async function setSubscriberBlocked(telegramUsername, enabled) {
    await request(`/admin/subscribers/${telegramUsername}/block`, {
      method: 'PATCH',
      body: { enabled },
    });
    updateAdminSubscriber(telegramUsername, (item) => ({
      isBlocked: enabled,
      notificationsEnabled: enabled ? false : item.notificationsEnabled,
    }));
  }

  async function setSubscriberAdmin(telegramUsername, enabled) {
    await request(`/admin/subscribers/${telegramUsername}/admin`, {
      method: 'PATCH',
      body: { enabled },
    });
    updateAdminSubscriber(telegramUsername, { isAdmin: enabled });
  }

  function clearError() {
    errorCode.value = '';
    errorStatus.value = 0;
  }

  function isCardTimelineItem(cardId) {
    return (item) => item.cardId === cardId && item.kind === 'card';
  }

  function removeTimelineItems(predicate) {
    timelineItems.value = timelineItems.value.filter(
      (item) => !predicate(item),
    );
  }

  function updateTimelineItemsByCard(cardId, patch) {
    timelineItems.value = timelineItems.value.map((item) =>
      item.cardId === cardId ? { ...item, ...resolvePatch(patch, item) } : item,
    );
  }

  function updateFeedItemsByCard(cardId, patch) {
    feedItems.value = feedItems.value.map((item) =>
      item.cardId === cardId ? { ...item, ...resolvePatch(patch, item) } : item,
    );
  }

  function updateActiveCard(cardId, patch) {
    if (activeCard.value?.id !== cardId) return;
    activeCard.value = {
      ...activeCard.value,
      ...resolvePatch(patch, activeCard.value),
    };
  }

  function updateActiveCardTask(taskId, patch) {
    if (!activeCard.value?.tasks) return;
    activeCard.value = {
      ...activeCard.value,
      tasks: activeCard.value.tasks.map((task) =>
        task.id === taskId ? { ...task, ...resolvePatch(patch, task) } : task,
      ),
    };
  }

  function updateFeedItem(id, patch) {
    feedItems.value = feedItems.value.map((item) =>
      item.id === id ? { ...item, ...resolvePatch(patch, item) } : item,
    );
  }

  function updateAdminSubscriber(telegramUsername, patch) {
    adminSubscribers.value = adminSubscribers.value.map((item) =>
      item.telegramUsername === telegramUsername
        ? { ...item, ...resolvePatch(patch, item) }
        : item,
    );
  }

  function resolvePatch(patch, item) {
    return typeof patch === 'function' ? patch(item) : patch;
  }

  function incrementCompleted(item) {
    if (!item) return;
    timelineStats.value = {
      ...timelineStats.value,
      completed: normalizeCount(timelineStats.value.completed) + 1,
    };
  }

  function normalizeCount(value) {
    const count = Number(value);
    return Number.isFinite(count) && count > 0 ? Math.round(count) : 0;
  }

  return {
    me,
    timelineItems,
    timelineStats,
    feedItems,
    adminSubscribers,
    activeCard,
    isLoading,
    errorCode,
    errorStatus,
    preferences,
    hasWorkspace,
    fetchMe,
    bindWorkspace,
    savePreferences,
    setMasterNotifications,
    setOnboardingCompleted,
    fetchTasks,
    fetchCard,
    completeCard,
    completeTask,
    addComment,
    muteCard,
    unmuteCard,
    fetchFeed,
    markFeedRead,
    fetchAdminSubscribers,
    setSubscriberBlocked,
    setSubscriberAdmin,
    clearError,
  };
});
