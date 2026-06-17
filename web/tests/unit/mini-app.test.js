import { afterEach, describe, expect, it, vi } from 'vitest';
import { DEFAULT_PREFERENCES } from '../../app/stores/mini-app.js';
import { mockTaskToCard } from '../../app/utils/mockTasks.mjs';

const importFresh = async (path) => {
  vi.resetModules();
  return import(path);
};

afterEach(() => {
  vi.doUnmock('pinia');
  vi.unstubAllGlobals();
  window.localStorage.clear();
  document.cookie = 'telegram_init_data=; path=/; max-age=0';
});

describe('mini app helpers', () => {
  it('uses Figma notification defaults', () => {
    expect(DEFAULT_PREFERENCES).toEqual({
      assignments: true,
      comments: true,
      changes: false,
      done: true,
    });
  });

  it('builds rich mock card details for Figma task screens', () => {
    const card = mockTaskToCard({
      id: 'mock:task:1',
      kind: 'task',
      cardId: 'mock-card-1',
      taskId: 'mock-task-1',
      title: 'Презентація сторінки послуг',
      boardName: 'Vovk Brand',
      dueAt: '2026-06-17T13:00:00.000Z',
      members: [
        { id: 'm1', username: 'yan', name: 'Ян', initials: 'Я' },
        { id: 'm2', username: 'julia', name: 'Юлія', initials: 'Ю' },
      ],
    });

    expect(card).toMatchObject({
      id: 'mock-card-1',
      title: 'Презентація сторінки послуг',
      boardName: 'Vovk Brand',
      cardWorkspaceUrl: '#',
    });
    expect(card.members).toHaveLength(2);
    expect(card.tasks).toHaveLength(4);
    expect(card.tasks[0]).toMatchObject({
      id: 'mock-task-1',
      isCompleted: false,
    });
    expect(card.description).toContain('Lorem ipsum');
    expect(card.comments).toHaveLength(3);
  });

  it('adds Telegram init data to API requests', async () => {
    const useRequest = vi.fn();

    vi.stubGlobal('useTelegramWebApp', () => ({
      initData: { value: 'auth_date=1&hash=test' },
    }));
    vi.stubGlobal('useMonitorDevAuth', () => ({
      getToken: () => '',
    }));
    vi.stubGlobal('useRequest', useRequest);

    const { useTelegramRequest } = await importFresh(
      '../../app/composables/useTelegramRequest.js',
    );

    useTelegramRequest('/me', {
      method: 'GET',
      headers: { 'X-Trace': 'test' },
    });

    expect(useRequest).toHaveBeenCalledWith(
      '/me',
      {
        method: 'GET',
        headers: {
          'X-Trace': 'test',
          'X-Telegram-Init-Data': 'auth_date=1&hash=test',
        },
      },
      'v1',
    );
  });

  it('adds monitor dev auth token to API requests', async () => {
    const useRequest = vi.fn();

    vi.stubGlobal('useTelegramWebApp', () => ({
      initData: { value: '' },
    }));
    vi.stubGlobal('useMonitorDevAuth', () => ({
      getToken: () => 'dev-secret',
    }));
    vi.stubGlobal('useRequest', useRequest);

    const { useTelegramRequest } = await importFresh(
      '../../app/composables/useTelegramRequest.js',
    );

    useTelegramRequest('/me');

    expect(useRequest).toHaveBeenCalledWith(
      '/me',
      {
        headers: {
          'X-Monitor-Dev-Auth': 'dev-secret',
        },
      },
      'v1',
    );
  });

  it('labels feed events from the current workspace user in second person', async () => {
    vi.stubGlobal('shallowRef', (value) => ({ value }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));
    vi.stubGlobal('watch', vi.fn());

    const { useMiniAppNavigation } = await importFresh(
      '../../app/composables/mini/useMiniAppNavigation.js',
    );

    const navigation = useMiniAppNavigation({
      activeTab: { value: 'feed' },
      clearTaskActionItem: vi.fn(),
      currentStep: { value: 'app' },
      loadTasks: vi.fn(),
      miniApp: {
        me: {
          workspaceUsername: 'yan',
          workspaceDisplayName: 'Yan',
        },
        adminSubscribers: [],
      },
      withAppError: vi.fn(),
    });

    expect(
      navigation.feedTitle({
        type: 'completeTask',
        actor: { username: 'yan', name: 'Yan' },
      }),
    ).toBe('Ви виконали задачу');
  });

  it('maps store actions to API endpoints and updates me', async () => {
    const responses = {
      '/me': {
        telegramUsername: 'yan',
        workspaceUsername: '',
        workspaceExists: false,
        notificationsEnabled: false,
        preferences: DEFAULT_PREFERENCES,
        needsStart: false,
        onboardingCompleted: false,
      },
      '/me/bind': {
        telegramUsername: 'yan',
        workspaceUsername: 'workspace_yan',
        workspaceExists: true,
        notificationsEnabled: false,
        preferences: DEFAULT_PREFERENCES,
        needsStart: false,
        onboardingCompleted: false,
      },
      '/me/preferences': {
        telegramUsername: 'yan',
        workspaceUsername: 'workspace_yan',
        workspaceExists: true,
        notificationsEnabled: false,
        preferences: {
          assignments: false,
          comments: true,
          changes: true,
          done: false,
        },
        needsStart: false,
        onboardingCompleted: false,
      },
      '/me/notifications': {
        telegramUsername: 'yan',
        workspaceUsername: 'workspace_yan',
        workspaceExists: true,
        notificationsEnabled: true,
        preferences: DEFAULT_PREFERENCES,
        needsStart: false,
        onboardingCompleted: false,
      },
      '/me/onboarding': {
        telegramUsername: 'yan',
        workspaceUsername: 'workspace_yan',
        workspaceExists: true,
        notificationsEnabled: true,
        preferences: DEFAULT_PREFERENCES,
        needsStart: false,
        onboardingCompleted: true,
      },
    };
    const useTelegramRequest = vi.fn(async (url) => ({
      _data: responses[url],
    }));

    vi.doMock('pinia', () => ({
      defineStore: (_name, setup) => () => setup(),
    }));
    vi.stubGlobal('shallowRef', (value) => ({ value }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));
    vi.stubGlobal('useTelegramRequest', useTelegramRequest);

    const { useMiniAppStore } = await importFresh(
      '../../app/stores/mini-app.js',
    );

    const store = useMiniAppStore();
    await store.fetchMe();
    await store.bindWorkspace('workspace_yan');
    await store.savePreferences({
      assignments: false,
      comments: true,
      changes: true,
      done: false,
    });
    await store.setMasterNotifications(true);
    await store.setOnboardingCompleted(true);

    expect(useTelegramRequest).toHaveBeenNthCalledWith(1, '/me', {});
    expect(useTelegramRequest).toHaveBeenNthCalledWith(2, '/me/bind', {
      method: 'POST',
      body: { workspaceUsername: 'workspace_yan' },
    });
    expect(useTelegramRequest).toHaveBeenNthCalledWith(3, '/me/preferences', {
      method: 'PATCH',
      body: {
        assignments: false,
        comments: true,
        changes: true,
        done: false,
      },
    });
    expect(useTelegramRequest).toHaveBeenNthCalledWith(4, '/me/notifications', {
      method: 'PATCH',
      body: { enabled: true },
    });
    expect(useTelegramRequest).toHaveBeenNthCalledWith(5, '/me/onboarding', {
      method: 'PATCH',
      body: { completed: true },
    });
    expect(store.me.value.notificationsEnabled).toBe(true);
    expect(store.me.value.onboardingCompleted).toBe(true);
  });

  it('routes workspace users with incomplete onboarding to welcome', async () => {
    let mounted;
    const currentStep = { value: 'loading' };
    const activeTab = { value: 'tasks' };
    const miniApp = {
      me: null,
      preferences: DEFAULT_PREFERENCES,
      hasWorkspace: true,
      errorCode: '',
      errorStatus: 0,
      fetchMe: vi.fn(async () => {
        miniApp.me = {
          telegramUsername: 'yan',
          workspaceUsername: 'yan',
          workspaceExists: true,
          notificationsEnabled: false,
          preferences: DEFAULT_PREFERENCES,
          needsStart: false,
          onboardingCompleted: false,
        };
      }),
      clearError: vi.fn(),
    };
    const loadTasks = vi.fn();

    vi.stubGlobal('shallowRef', (value) => ({ value }));
    vi.stubGlobal('reactive', (value) => ({ ...value }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));
    vi.stubGlobal('onMounted', (fn) => {
      mounted = fn();
    });

    const { useMiniAppSession } = await importFresh(
      '../../app/composables/mini/useMiniAppSession.js',
    );

    useMiniAppSession({
      activeTab,
      currentStep,
      devAuth: {
        hasToken: { value: true },
        syncFromURL: vi.fn(),
      },
      loadTasks,
      miniApp,
      telegram: {
        initData: { value: '' },
        init: vi.fn(),
        ready: vi.fn(),
        expand: vi.fn(),
        waitForInitData: vi.fn(),
      },
    });
    await mounted;

    expect(currentStep.value).toBe('welcome');
    expect(loadTasks).not.toHaveBeenCalled();
  });

  it('enters app only after onboarding is completed', async () => {
    let mounted;
    const currentStep = { value: 'loading' };
    const activeTab = { value: 'tasks' };
    const miniApp = {
      me: null,
      preferences: DEFAULT_PREFERENCES,
      hasWorkspace: true,
      errorCode: '',
      errorStatus: 0,
      fetchMe: vi.fn(async () => {
        miniApp.me = {
          telegramUsername: 'yan',
          workspaceUsername: 'yan',
          workspaceExists: true,
          notificationsEnabled: true,
          preferences: DEFAULT_PREFERENCES,
          needsStart: false,
          onboardingCompleted: true,
        };
      }),
      clearError: vi.fn(),
    };
    const loadTasks = vi.fn();

    vi.stubGlobal('shallowRef', (value) => ({ value }));
    vi.stubGlobal('reactive', (value) => ({ ...value }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));
    vi.stubGlobal('onMounted', (fn) => {
      mounted = fn();
    });

    const { useMiniAppSession } = await importFresh(
      '../../app/composables/mini/useMiniAppSession.js',
    );

    useMiniAppSession({
      activeTab,
      currentStep,
      devAuth: {
        hasToken: { value: true },
        syncFromURL: vi.fn(),
      },
      loadTasks,
      miniApp,
      telegram: {
        initData: { value: '' },
        init: vi.fn(),
        ready: vi.fn(),
        expand: vi.fn(),
        waitForInitData: vi.fn(),
      },
    });
    await mounted;

    expect(currentStep.value).toBe('app');
    expect(loadTasks).toHaveBeenCalledOnce();
  });

  it('switches workspace binding between suggested and manual modes', async () => {
    const currentStep = { value: 'welcome' };
    const miniApp = {
      me: {
        telegramUsername: 'netherg',
        workspaceUsername: 'netherg',
        workspaceExists: true,
        workspaceDisplayName: 'Лозов Миколай',
        notificationsEnabled: false,
        preferences: DEFAULT_PREFERENCES,
      },
      preferences: DEFAULT_PREFERENCES,
      errorCode: '',
      errorStatus: 0,
      clearError: vi.fn(),
      bindWorkspace: vi.fn(async () => {}),
    };

    vi.stubGlobal('shallowRef', (value) => ({ value }));
    vi.stubGlobal('reactive', (value) => ({ ...value }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));
    vi.stubGlobal('onMounted', vi.fn());

    const { useMiniAppSession } = await importFresh(
      '../../app/composables/mini/useMiniAppSession.js',
    );

    const session = useMiniAppSession({
      activeTab: { value: 'tasks' },
      currentStep,
      devAuth: {
        hasToken: { value: true },
        syncFromURL: vi.fn(),
      },
      loadTasks: vi.fn(),
      miniApp,
      telegram: {
        initData: { value: '' },
        init: vi.fn(),
        ready: vi.fn(),
        expand: vi.fn(),
        waitForInitData: vi.fn(),
      },
    });

    session.openBindStep();
    expect(session.selectedWorkspaceMode.value).toBe('suggested');
    expect(session.workspaceInput.value).toBe('netherg');

    session.selectManualWorkspace();
    expect(session.selectedWorkspaceMode.value).toBe('manual');
    expect(session.workspaceInput.value).toBe('');

    session.updateWorkspaceInput('other-user');
    expect(session.selectedWorkspaceMode.value).toBe('manual');
    expect(session.workspaceInput.value).toBe('other-user');

    session.selectSuggestedWorkspace();
    expect(session.selectedWorkspaceMode.value).toBe('suggested');
    expect(session.workspaceInput.value).toBe('netherg');

    await session.confirmBinding();
    expect(miniApp.bindWorkspace).toHaveBeenCalledWith('netherg');
  });

  it('maps timeline, feed and admin actions to monitor API endpoints', async () => {
    const responses = {
      '/tasks': {
        stats: { completed: 2 },
        items: [
          { id: 'card:c1', kind: 'card', cardId: 'c1', title: 'Card' },
          { id: 'task:t1', kind: 'task', cardId: 'c1', taskId: 't1' },
        ],
      },
      '/cards/c1': {
        id: 'c1',
        title: 'Card',
        tasks: [{ id: 't1', isCompleted: false }],
      },
      '/cards/c1/complete': {},
      '/tasks/t1/complete': {},
      '/cards/c1/comments': {},
      '/cards/c1/mute': {},
      '/feed': { items: [{ id: 'n:1', isRead: false, cardId: 'c1' }] },
      '/feed/n:1/read': {},
      '/admin/subscribers': {
        items: [
          {
            telegramUsername: 'yan',
            isAdmin: false,
            isBlocked: false,
            notificationsEnabled: true,
          },
        ],
      },
      '/admin/subscribers/yan/block': {},
      '/admin/subscribers/yan/admin': {},
    };
    const useTelegramRequest = vi.fn(async (url) => ({
      _data: responses[url],
    }));

    vi.doMock('pinia', () => ({
      defineStore: (_name, setup) => () => setup(),
    }));
    vi.stubGlobal('shallowRef', (value) => ({ value }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));
    vi.stubGlobal('useTelegramRequest', useTelegramRequest);

    const { useMiniAppStore } = await importFresh(
      '../../app/stores/mini-app.js',
    );

    const store = useMiniAppStore();
    await store.fetchTasks({ view: 'list' });
    await store.fetchCard('c1');
    await store.completeTask('t1');
    expect(store.activeCard.value.tasks[0].isCompleted).toBe(true);
    await store.completeCard('c1');
    await store.addComment('c1', 'Привіт');
    await store.muteCard('c1');
    await store.unmuteCard('c1');
    await store.fetchFeed({ type: 'comments' });
    await store.markFeedRead('n:1');
    await store.fetchAdminSubscribers({ status: 'active' });
    await store.setSubscriberBlocked('yan', true);
    await store.setSubscriberAdmin('yan', true);

    expect(useTelegramRequest).toHaveBeenNthCalledWith(1, '/tasks', {
      query: { view: 'list' },
    });
    expect(useTelegramRequest).toHaveBeenNthCalledWith(2, '/cards/c1', {});
    expect(useTelegramRequest).toHaveBeenNthCalledWith(
      3,
      '/tasks/t1/complete',
      { method: 'POST' },
    );
    expect(useTelegramRequest).toHaveBeenNthCalledWith(
      4,
      '/cards/c1/complete',
      { method: 'POST' },
    );
    expect(useTelegramRequest).toHaveBeenNthCalledWith(
      5,
      '/cards/c1/comments',
      {
        method: 'POST',
        body: { text: 'Привіт' },
      },
    );
    expect(useTelegramRequest).toHaveBeenNthCalledWith(6, '/cards/c1', {});
    expect(useTelegramRequest).toHaveBeenNthCalledWith(7, '/cards/c1/mute', {
      method: 'POST',
    });
    expect(useTelegramRequest).toHaveBeenNthCalledWith(8, '/cards/c1/mute', {
      method: 'DELETE',
    });
    expect(useTelegramRequest).toHaveBeenNthCalledWith(9, '/feed', {
      query: { type: 'comments' },
    });
    expect(useTelegramRequest).toHaveBeenNthCalledWith(10, '/feed/n:1/read', {
      method: 'PATCH',
    });
    expect(useTelegramRequest).toHaveBeenNthCalledWith(
      11,
      '/admin/subscribers',
      {
        query: { status: 'active' },
      },
    );
    expect(useTelegramRequest).toHaveBeenNthCalledWith(
      12,
      '/admin/subscribers/yan/block',
      {
        method: 'PATCH',
        body: { enabled: true },
      },
    );
    expect(useTelegramRequest).toHaveBeenNthCalledWith(
      13,
      '/admin/subscribers/yan/admin',
      {
        method: 'PATCH',
        body: { enabled: true },
      },
    );
    expect(store.activeCard.value.isMuted).toBe(false);
    expect(store.feedItems.value[0].isRead).toBe(true);
    expect(store.timelineStats.value.completed).toBe(4);
    expect(store.adminSubscribers.value[0]).toMatchObject({
      isAdmin: true,
      isBlocked: true,
      notificationsEnabled: false,
    });
  });

  it('fills sparse timelines with local mock tasks for design checks', async () => {
    const useTelegramRequest = vi.fn(async () => ({
      _data: {
        items: [
          {
            id: 'card:c1',
            kind: 'card',
            cardId: 'c1',
            title: 'Card',
            boardName: 'Board',
            dueAt: '2026-06-17T09:00:00.000Z',
          },
        ],
      },
    }));

    vi.doMock('pinia', () => ({
      defineStore: (_name, setup) => () => setup(),
    }));
    vi.stubGlobal('shallowRef', (value) => ({ value }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));
    vi.stubGlobal('useTelegramRequest', useTelegramRequest);

    const { useMiniAppStore } = await importFresh(
      '../../app/stores/mini-app.js',
    );

    const store = useMiniAppStore();
    await store.fetchTasks({
      from: '2026-06-15T00:00:00.000Z',
      to: '2026-06-21T23:59:59.999Z',
      view: 'calendar',
      mockTasks: true,
      minMockTasks: 6,
      mockSelectedDate: '2026-06-17',
    });

    expect(useTelegramRequest).toHaveBeenCalledWith('/tasks', {
      query: {
        from: '2026-06-15T00:00:00.000Z',
        to: '2026-06-21T23:59:59.999Z',
        view: 'calendar',
      },
    });
    expect(store.timelineItems.value).toHaveLength(6);
    expect(store.timelineItems.value.some((item) => item.isMock)).toBe(true);
  });

  it('updates task mute state optimistically and clears pending state', async () => {
    let resolveMute;
    const muteRequest = new Promise((resolve) => {
      resolveMute = resolve;
    });
    const miniApp = {
      activeCard: null,
      timelineItems: [{ cardId: 'c1', isMuted: false, title: 'Card' }],
      muteCard: vi.fn(() => muteRequest),
      unmuteCard: vi.fn(),
    };
    const withAppError = vi.fn(async (_message, action) => {
      await action();
      return true;
    });

    vi.stubGlobal('shallowRef', (value) => ({ value }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));

    const { useMiniTaskTimeline } = await importFresh(
      '../../app/composables/mini/useMiniTaskTimeline.js',
    );
    const timeline = useMiniTaskTimeline({
      activeTab: { value: 'tasks' },
      displayName: { value: 'Yan' },
      miniApp,
      route: { query: {} },
      runtimeConfig: { public: {} },
      searchQuery: { value: '' },
      taskView: { value: 'list' },
      withAppError,
    });

    timeline.openTaskMenu(miniApp.timelineItems[0]);
    const pending = timeline.muteTaskActionItem(miniApp.timelineItems[0]);

    expect(miniApp.timelineItems[0].isMuted).toBe(true);
    expect(timeline.taskActionItem.value).toMatchObject({
      isMuted: true,
      isMutePending: true,
    });

    resolveMute();
    await pending;

    expect(miniApp.muteCard).toHaveBeenCalledWith('c1');
    expect(timeline.taskActionItem.value).toMatchObject({
      isMuted: true,
      isMutePending: false,
    });
  });

  it('updates feed mute state optimistically and clears pending state', async () => {
    let resolveUnmute;
    const unmuteRequest = new Promise((resolve) => {
      resolveUnmute = resolve;
    });
    const miniApp = {
      activeCard: null,
      feedItems: [{ cardId: 'c1', isMuted: true, cardTitle: 'Card' }],
      timelineItems: [{ cardId: 'c1', isMuted: true, title: 'Card' }],
      muteCard: vi.fn(),
      unmuteCard: vi.fn(() => unmuteRequest),
    };
    const withAppError = vi.fn(async (_message, action) => {
      await action();
      return true;
    });

    vi.stubGlobal('shallowRef', (value) => ({ value }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));

    const { useMiniTaskTimeline } = await importFresh(
      '../../app/composables/mini/useMiniTaskTimeline.js',
    );
    const timeline = useMiniTaskTimeline({
      activeTab: { value: 'feed' },
      displayName: { value: 'Yan' },
      miniApp,
      route: { query: {} },
      runtimeConfig: { public: {} },
      searchQuery: { value: '' },
      taskView: { value: 'list' },
      withAppError,
    });

    timeline.openFeedMenu(miniApp.feedItems[0]);
    const pending = timeline.muteFeedActionItem(miniApp.feedItems[0]);

    expect(miniApp.feedItems[0].isMuted).toBe(false);
    expect(miniApp.timelineItems[0].isMuted).toBe(false);
    expect(timeline.feedActionItem.value).toMatchObject({
      isMuted: false,
      isMutePending: true,
    });

    resolveUnmute();
    await pending;

    expect(miniApp.unmuteCard).toHaveBeenCalledWith('c1');
    expect(timeline.feedActionItem.value).toMatchObject({
      isMuted: false,
      isMutePending: false,
    });
  });

  it('rolls optimistic task mute state back when the request fails', async () => {
    const miniApp = {
      activeCard: null,
      timelineItems: [{ cardId: 'c1', isMuted: false, title: 'Card' }],
      muteCard: vi.fn(async () => {
        throw new Error('failed');
      }),
      unmuteCard: vi.fn(),
    };
    const withAppError = vi.fn(async (_message, action) => {
      try {
        await action();
        return true;
      } catch {
        return false;
      }
    });

    vi.stubGlobal('shallowRef', (value) => ({ value }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));

    const { useMiniTaskTimeline } = await importFresh(
      '../../app/composables/mini/useMiniTaskTimeline.js',
    );
    const timeline = useMiniTaskTimeline({
      activeTab: { value: 'tasks' },
      displayName: { value: 'Yan' },
      miniApp,
      route: { query: {} },
      runtimeConfig: { public: {} },
      searchQuery: { value: '' },
      taskView: { value: 'list' },
      withAppError,
    });

    timeline.openTaskMenu(miniApp.timelineItems[0]);
    await timeline.muteTaskActionItem(miniApp.timelineItems[0]);

    expect(miniApp.timelineItems[0].isMuted).toBe(false);
    expect(timeline.taskActionItem.value).toMatchObject({
      isMuted: false,
      isMutePending: false,
    });
  });
});
