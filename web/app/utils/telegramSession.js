export const TELEGRAM_SESSION_COOKIE = 'telegram_init_data';
export const TELEGRAM_SESSION_STORAGE = 'telegramInitData';
export const TELEGRAM_SESSION_MAX_AGE = 60 * 60 * 24;

export function normalizeTelegramInitData(value) {
  const raw = String(value || '').trim();
  if (!raw) return '';

  if (raw.includes('=') || !raw.includes('%')) return raw;

  try {
    return decodeURIComponent(raw);
  } catch {
    return raw;
  }
}

export function isCompleteTelegramInitData(value) {
  const initData = normalizeTelegramInitData(value);
  if (!initData) return false;

  const params = new URLSearchParams(initData);

  return Boolean(
    params.get('auth_date') && params.get('hash') && params.get('user'),
  );
}

export function readCookieValue(cookieString, name) {
  const cookie = String(cookieString || '')
    .split(';')
    .map((part) => part.trim())
    .find((part) => part.startsWith(`${name}=`));

  if (!cookie) return '';

  const value = cookie.split('=').slice(1).join('=');

  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}

export function buildSessionCookie(
  initData,
  maxAge = TELEGRAM_SESSION_MAX_AGE,
) {
  return `${TELEGRAM_SESSION_COOKIE}=${encodeURIComponent(
    normalizeTelegramInitData(initData),
  )}; path=/; max-age=${maxAge}; SameSite=Lax`;
}

export function readTelegramSessionInitData({ cookieString, storage } = {}) {
  const cookieInitData = readCookieValue(cookieString, TELEGRAM_SESSION_COOKIE);

  if (isCompleteTelegramInitData(cookieInitData)) {
    return {
      initData: normalizeTelegramInitData(cookieInitData),
      source: 'telegram_init_data cookie',
    };
  }

  let storedInitData = '';

  try {
    storedInitData = storage?.getItem?.(TELEGRAM_SESSION_STORAGE) || '';
  } catch {
    storedInitData = '';
  }

  if (isCompleteTelegramInitData(storedInitData)) {
    return {
      initData: normalizeTelegramInitData(storedInitData),
      source: 'telegram_init_data storage',
    };
  }

  return { initData: '', source: '' };
}

export function resolveTelegramInitData(candidates, storedSession) {
  const prepared = [
    ...(candidates || []),
    storedSession || { initData: '', source: '' },
  ]
    .map((candidate) => ({
      initData: normalizeTelegramInitData(candidate?.initData),
      source: candidate?.source || '',
    }))
    .filter((candidate) => candidate.initData);

  return (
    prepared.find((candidate) =>
      isCompleteTelegramInitData(candidate.initData),
    ) ||
    prepared[0] || { initData: '', source: '' }
  );
}

export function persistTelegramSessionInitData(
  initData,
  { documentRef, storage, maxAge = TELEGRAM_SESSION_MAX_AGE } = {},
) {
  if (!isCompleteTelegramInitData(initData)) return false;

  const normalized = normalizeTelegramInitData(initData);

  try {
    storage?.setItem?.(TELEGRAM_SESSION_STORAGE, normalized);
  } catch {
    // Cookie fallback is enough for the mini-app session.
  }

  if (documentRef) {
    documentRef.cookie = buildSessionCookie(normalized, maxAge);
  }

  return true;
}
