const DEV_AUTH_COOKIE = 'monitor_dev_auth';
const DEV_AUTH_STORAGE = 'monitorDevAuthToken';
const DEV_AUTH_BOOTSTRAP = '__MONITOR_DEV_AUTH_TOKEN__';
const EXPIRED_DEV_AUTH_COOKIE = [
  `${DEV_AUTH_COOKIE}=`,
  'path=/',
  'max-age=0',
  'expires=Thu, 01 Jan 1970 00:00:00 GMT',
  'SameSite=Lax',
].join('; ');
const DEV_AUTH_PARAMS = [
  'monitorDevToken',
  'monitorDevAuth',
  'monitor_dev_auth',
  'devAuth',
];

export const useMonitorDevAuth = () => {
  const token = useState('monitor-dev-auth-token', () => '');
  const hasToken = computed(() => Boolean(token.value));

  function clearStoredToken() {
    if (typeof window === 'undefined' || typeof document === 'undefined') {
      return;
    }

    try {
      window.localStorage.removeItem(DEV_AUTH_STORAGE);
    } catch {
      // Ignore blocked storage; cookie cleanup below is enough for auth.
    }

    document.cookie = EXPIRED_DEV_AUTH_COOKIE;
  }

  function consumeBootstrapToken() {
    if (typeof window === 'undefined') return '';

    const value =
      typeof window[DEV_AUTH_BOOTSTRAP] === 'string'
        ? window[DEV_AUTH_BOOTSTRAP]
        : '';
    delete window[DEV_AUTH_BOOTSTRAP];
    return value.trim();
  }

  function setToken(nextToken) {
    const normalized = nextToken.trim();
    token.value = normalized;
    clearStoredToken();
  }

  function syncFromURL() {
    if (typeof window === 'undefined') return;

    const url = new URL(window.location.href);
    const urlToken = DEV_AUTH_PARAMS.map((key) => url.searchParams.get(key))
      .find(Boolean)
      ?.trim();
    const nextToken = urlToken || consumeBootstrapToken();

    if (nextToken) {
      setToken(nextToken);
      for (const key of DEV_AUTH_PARAMS) {
        url.searchParams.delete(key);
      }
      window.history.replaceState(
        {},
        '',
        `${url.pathname}${url.search}${url.hash}`,
      );
      return;
    }

    setToken('');
  }

  function getToken() {
    return token.value;
  }

  return {
    token,
    hasToken,
    getToken,
    setToken,
    syncFromURL,
  };
};
