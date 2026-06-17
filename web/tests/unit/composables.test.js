import { afterEach, describe, expect, it, vi } from 'vitest';

const importFresh = async (path) => {
  vi.resetModules();
  return import(path);
};

afterEach(() => {
  vi.doUnmock('@/stores/global');
  vi.doUnmock('@fluejs/noscroll');
  vi.unstubAllGlobals();
  window.localStorage.clear();
  delete window.__MONITOR_DEV_AUTH_TOKEN__;
  document.cookie = 'monitor_dev_auth=; path=/; max-age=0';
  document.cookie = 'telegram_init_data=; path=/; max-age=0';
});

describe('composables', () => {
  it('passes the shared breakpoint map to VueUse', async () => {
    const useBreakpoints = vi.fn((breakpoints) => ({ breakpoints }));

    vi.stubGlobal('useBreakpoints', useBreakpoints);

    const { useCustomBreakpoints } = await importFresh(
      '../../app/composables/useCustomBreakpoints.js',
    );

    expect(useCustomBreakpoints()).toEqual({
      breakpoints: {
        xs: 375,
        sm: 768,
        md: 1024,
        lg: 1280,
        xl: 1600,
        xxl: 1920,
      },
    });
    expect(useBreakpoints).toHaveBeenCalledTimes(1);
  });

  it('returns the current request host', async () => {
    vi.stubGlobal('useRequestURL', () => ({ host: 'example.test' }));

    const { default: useIsDomain } = await importFresh(
      '../../app/composables/useIsDomain.js',
    );

    expect(useIsDomain()).toBe('example.test');
  });

  it('builds API requests with query serialization and runtime base fallback', async () => {
    const raw = vi.fn().mockResolvedValue({ ok: true });

    vi.stubGlobal('$fetch', { raw });
    vi.stubGlobal('useRuntimeConfig', () => ({
      app: { baseURL: '/site' },
      public: { cmsUrl: '' },
    }));

    const { useRequest } = await importFresh(
      '../../app/composables/useRequest.js',
    );

    await useRequest(
      '/properties',
      {
        method: 'GET',
        query: { rooms: ['studio', '1br'], empty: null },
      },
      'v1',
    );

    expect(raw).toHaveBeenCalledWith(
      '/site/api/v1/properties?rooms[]=studio&rooms[]=1br',
      expect.objectContaining({
        retry: 1,
        retryDelay: 500,
        timeout: 30000,
        method: 'GET',
        query: undefined,
      }),
    );
  });

  it('uses monitor dev auth from URL only for the current page session', async () => {
    window.history.pushState({}, '', '/?monitorDevToken=dev-secret&tab=tasks');
    window.localStorage.setItem('monitorDevAuthToken', 'stale-secret');
    document.cookie = 'monitor_dev_auth=stale-secret; path=/';
    vi.stubGlobal('useState', (_key, init) => ({ value: init() }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));

    const { useMonitorDevAuth } = await importFresh(
      '../../app/composables/useMonitorDevAuth.js',
    );

    const devAuth = useMonitorDevAuth();
    devAuth.syncFromURL();

    expect(devAuth.getToken()).toBe('dev-secret');
    expect(window.localStorage.getItem('monitorDevAuthToken')).toBeNull();
    expect(document.cookie).not.toContain('monitor_dev_auth=');
    expect(window.location.search).toBe('?tab=tasks');
  });

  it('clears stale monitor dev auth when URL has no token', async () => {
    window.history.pushState({}, '', '/?tab=tasks');
    window.localStorage.setItem('monitorDevAuthToken', 'stale-secret');
    document.cookie = 'monitor_dev_auth=stale-secret; path=/';
    vi.stubGlobal('useState', (_key, init) => ({ value: init() }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));

    const { useMonitorDevAuth } = await importFresh(
      '../../app/composables/useMonitorDevAuth.js',
    );

    const devAuth = useMonitorDevAuth();
    devAuth.syncFromURL();

    expect(devAuth.getToken()).toBe('');
    expect(devAuth.hasToken.value).toBe(false);
    expect(window.localStorage.getItem('monitorDevAuthToken')).toBeNull();
    expect(document.cookie).not.toContain('monitor_dev_auth=');
  });

  it('uses monitor dev auth from the production bootstrap once', async () => {
    window.history.pushState({}, '', '/?tab=tasks');
    window.__MONITOR_DEV_AUTH_TOKEN__ = 'dev-secret';
    vi.stubGlobal('useState', (_key, init) => ({ value: init() }));
    vi.stubGlobal('computed', (getter) => ({
      get value() {
        return getter();
      },
    }));

    const { useMonitorDevAuth } = await importFresh(
      '../../app/composables/useMonitorDevAuth.js',
    );

    const devAuth = useMonitorDevAuth();
    devAuth.syncFromURL();

    expect(devAuth.getToken()).toBe('dev-secret');
    expect(window.__MONITOR_DEV_AUTH_TOKEN__).toBeUndefined();
  });

  it('stores complete Telegram init data as a mini-app session', async () => {
    const telegramInitData =
      'query_id=q&user=%7B%22id%22%3A1%2C%22username%22%3A%22netherg%22%7D&auth_date=1710000000&hash=abc';

    const {
      persistTelegramSessionInitData,
      readTelegramSessionInitData,
      TELEGRAM_SESSION_STORAGE,
    } = await importFresh('../../app/utils/telegramSession.js');

    expect(
      persistTelegramSessionInitData(telegramInitData, {
        documentRef: document,
        storage: window.localStorage,
      }),
    ).toBe(true);

    expect(window.localStorage.getItem(TELEGRAM_SESSION_STORAGE)).toBe(
      telegramInitData,
    );
    expect(document.cookie).toContain('telegram_init_data=');
    expect(
      readTelegramSessionInitData({
        cookieString: document.cookie,
        storage: window.localStorage,
      }),
    ).toMatchObject({
      initData: telegramInitData,
      source: 'telegram_init_data cookie',
    });
  });

  it('prefers stored Telegram session over incomplete refresh payload', async () => {
    const telegramInitData =
      'query_id=q&user=%7B%22id%22%3A1%2C%22username%22%3A%22netherg%22%7D&auth_date=1710000000&hash=abc';

    const {
      readTelegramSessionInitData,
      resolveTelegramInitData,
      buildSessionCookie,
    } = await importFresh('../../app/utils/telegramSession.js');

    document.cookie = buildSessionCookie(telegramInitData);

    const resolved = resolveTelegramInitData(
      [{ initData: 'query_id', source: 'Telegram.WebApp.initData' }],
      readTelegramSessionInitData({
        cookieString: document.cookie,
        storage: window.localStorage,
      }),
    );

    expect(resolved).toEqual({
      initData: telegramInitData,
      source: 'telegram_init_data cookie',
    });
  });

  it('calculates route params with or without locales', async () => {
    vi.doMock('@/stores/global', () => ({
      useGlobalStore: () => ({}),
    }));
    vi.stubGlobal('useRoute', () => ({
      params: { all: ['projects', 'demo-project'] },
    }));

    const { extractSlugFromUrl, useRouteParams } = await importFresh(
      '../../app/composables/useRouteParams.js',
    );

    expect(extractSlugFromUrl('/projects/demo-project')).toBe('demo-project');
    expect(useRouteParams()).toMatchObject({
      lang: '',
      page: 'projects',
      slug: 'demo-project',
    });
  });

  it('recognizes locale-prefixed route params', async () => {
    vi.doMock('@/stores/global', () => ({
      useGlobalStore: () => ({
        languages: [{ code: 'en', isDefault: true }, { code: 'uk' }],
      }),
    }));
    vi.stubGlobal('useRoute', () => ({
      params: { all: ['uk', 'news', 'launch'] },
    }));

    const { useRouteParams } = await importFresh(
      '../../app/composables/useRouteParams.js',
    );

    expect(useRouteParams()).toMatchObject({
      lang: 'uk',
      page: 'news',
      slug: 'launch',
    });
  });

  it('exposes scroll lock package adapters', async () => {
    const disablePageScroll = vi.fn();
    const enablePageScroll = vi.fn();
    const updateAllScrollbarWidthAdjustment = vi.fn();

    vi.doMock('@fluejs/noscroll', () => ({
      disablePageScroll,
      enablePageScroll,
      updateAllScrollbarWidthAdjustment,
    }));

    const { useScrollLock } = await importFresh(
      '../../app/composables/useScrollLock.js',
    );
    const scrollLock = useScrollLock();

    scrollLock.lock();
    scrollLock.unlock();
    scrollLock.fillGaps();

    expect(disablePageScroll).toHaveBeenCalledTimes(1);
    expect(enablePageScroll).toHaveBeenCalledTimes(1);
    expect(updateAllScrollbarWidthAdjustment).toHaveBeenCalledTimes(1);
    expect(scrollLock.getScrollBarWidth()).toBe(0);
  });
});
