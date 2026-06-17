import {
  isCompleteTelegramInitData,
  persistTelegramSessionInitData,
  readTelegramSessionInitData,
  resolveTelegramInitData,
} from '../utils/telegramSession.js';

export const useTelegramWebApp = () => {
  const webApp = shallowRef(null);
  const initData = useState('telegram-init-data', () => '');
  const platform = useState('telegram-platform', () => '');
  const colorScheme = useState('telegram-color-scheme', () => 'dark');
  const isReady = useState('telegram-ready', () => false);
  const diagnostics = useState('telegram-diagnostics', () => ({
    hasWebApp: false,
    initDataLength: 0,
    initDataUnsafeUser: false,
    launchSource: '',
    locationHasTgWebAppData: false,
    platform: '',
    version: '',
  }));
  const listenersAttached = useState(
    'telegram-listeners-attached',
    () => false,
  );

  const isTelegram = computed(() => isCompleteTelegramInitData(initData.value));

  function getLaunchParamsFromLocation() {
    if (!import.meta.client) {
      return { initData: '', hasTgWebAppData: false };
    }

    const sources = [window.location.hash, window.location.search];

    for (const source of sources) {
      if (!source) continue;

      const queryStart = source.indexOf('?');
      const raw =
        queryStart >= 0
          ? source.slice(queryStart + 1)
          : source.replace(/^#/, '');
      const params = new URLSearchParams(raw);
      const rawInitData = params.get('tgWebAppData');

      if (rawInitData) {
        return { initData: rawInitData, hasTgWebAppData: true };
      }
    }

    return { initData: '', hasTgWebAppData: false };
  }

  function readStoredSession() {
    if (!import.meta.client) return { initData: '', source: '' };

    return readTelegramSessionInitData({
      cookieString: document.cookie,
      storage: window.localStorage,
    });
  }

  function persistSession(nextInitData) {
    if (!import.meta.client) return;

    persistTelegramSessionInitData(nextInitData, {
      documentRef: document,
      storage: window.localStorage,
    });
  }

  function syncDiagnostics(app, selectedSession, hasTgWebAppData) {
    if (!import.meta.client) return;

    diagnostics.value = {
      hasWebApp: Boolean(app),
      initDataLength: initData.value.length,
      initDataUnsafeUser: Boolean(app?.initDataUnsafe?.user),
      launchSource: selectedSession.source || '',
      locationHasTgWebAppData: hasTgWebAppData,
      platform: app?.platform || '',
      version: app?.version || '',
    };
  }

  function syncSafeArea(app) {
    if (!import.meta.client || !app) return;

    const root = document.documentElement;
    const safeArea = app.safeAreaInset || {};
    const contentSafeArea = app.contentSafeAreaInset || {};

    root.style.setProperty('--tg-safe-top', `${safeArea.top || 0}px`);
    root.style.setProperty('--tg-safe-right', `${safeArea.right || 0}px`);
    root.style.setProperty('--tg-safe-bottom', `${safeArea.bottom || 0}px`);
    root.style.setProperty('--tg-safe-left', `${safeArea.left || 0}px`);
    root.style.setProperty(
      '--tg-content-safe-top',
      `${contentSafeArea.top || 0}px`,
    );
    root.style.setProperty(
      '--tg-content-safe-bottom',
      `${contentSafeArea.bottom || 0}px`,
    );
    root.style.setProperty(
      '--tg-viewport-height',
      `${app.viewportHeight || window.innerHeight}px`,
    );
    root.style.setProperty(
      '--tg-viewport-stable-height',
      `${app.viewportStableHeight || app.viewportHeight || window.innerHeight}px`,
    );
  }

  function init() {
    if (!import.meta.client) return null;

    const app = window.Telegram?.WebApp;
    const { initData: fallbackInitData, hasTgWebAppData } =
      getLaunchParamsFromLocation();
    const storedSession = readStoredSession();

    if (!app) {
      const selectedSession = resolveTelegramInitData(
        [{ initData: fallbackInitData, source: 'tgWebAppData' }],
        storedSession,
      );

      initData.value = selectedSession.initData;
      persistSession(initData.value);
      syncDiagnostics(null, selectedSession, hasTgWebAppData);
      return null;
    }

    webApp.value = markRaw(app);
    const selectedSession = resolveTelegramInitData(
      [
        { initData: app.initData, source: 'Telegram.WebApp.initData' },
        { initData: fallbackInitData, source: 'tgWebAppData' },
      ],
      storedSession,
    );

    initData.value = selectedSession.initData;
    persistSession(initData.value);
    platform.value = app.platform || '';
    colorScheme.value = app.colorScheme || 'dark';
    syncSafeArea(app);
    syncDiagnostics(app, selectedSession, hasTgWebAppData);

    if (!listenersAttached.value) {
      listenersAttached.value = true;
      app.onEvent?.('safeAreaChanged', () => syncSafeArea(app));
      app.onEvent?.('contentSafeAreaChanged', () => syncSafeArea(app));
      app.onEvent?.('viewportChanged', () => syncSafeArea(app));
      app.onEvent?.('themeChanged', () => {
        colorScheme.value = app.colorScheme || 'dark';
      });
    }

    return app;
  }

  async function waitForInitData(timeout = 3000) {
    if (!import.meta.client) return null;

    const startedAt = Date.now();

    while (Date.now() - startedAt < timeout) {
      const app = init();

      if (isCompleteTelegramInitData(initData.value)) {
        return app;
      }

      await new Promise((resolve) => setTimeout(resolve, 50));
    }

    return init();
  }

  function ready() {
    const app = webApp.value || init();
    app?.ready?.();
    isReady.value = true;
  }

  function expand() {
    const app = webApp.value || init();
    app?.expand?.();
  }

  function triggerHapticImpact(style = 'light') {
    const app = webApp.value || init();

    try {
      app?.HapticFeedback?.impactOccurred?.(style);
    } catch {
      // Haptic feedback is optional in Telegram clients.
    }
  }

  return {
    webApp,
    initData,
    platform,
    colorScheme,
    diagnostics,
    isReady,
    isTelegram,
    init,
    waitForInitData,
    ready,
    expand,
    triggerHapticImpact,
  };
};
