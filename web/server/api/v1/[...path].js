import {
  isCompleteTelegramInitData,
  TELEGRAM_SESSION_COOKIE,
} from '../../../app/utils/telegramSession.js';

const DEV_AUTH_COOKIE = 'monitor_dev_auth';

export default defineEventHandler(async (event) => {
  const runtimeConfig = useRuntimeConfig(event);
  const monitorApiUrl =
    process.env.MONITOR_API_URL ||
    runtimeConfig.monitorApiUrl ||
    'http://localhost:8080';

  const rawPath = event.context.params?.path || '';
  const path = Array.isArray(rawPath) ? rawPath.join('/') : rawPath;
  const method = getMethod(event);
  const target = `${String(monitorApiUrl).replace(/\/+$/, '')}/api/v1/${path}`;
  const isAvatarRequest = path.startsWith('avatars/');
  const incomingHeaders = getRequestHeaders(event);
  const headers = {};

  const incomingInitData = incomingHeaders['x-telegram-init-data'] || '';
  const sessionInitData = getCookie(event, TELEGRAM_SESSION_COOKIE) || '';
  const telegramInitData = isCompleteTelegramInitData(incomingInitData)
    ? incomingInitData
    : isCompleteTelegramInitData(sessionInitData)
      ? sessionInitData
      : incomingInitData;

  if (telegramInitData) {
    headers['X-Telegram-Init-Data'] = telegramInitData;
  }
  const devAuthToken =
    incomingHeaders['x-monitor-dev-auth'] || getCookie(event, DEV_AUTH_COOKIE);
  if (devAuthToken) {
    headers['X-Monitor-Dev-Auth'] = devAuthToken;
  }
  if (incomingHeaders['content-type']) {
    headers['Content-Type'] = incomingHeaders['content-type'];
  }

  const body = ['GET', 'HEAD'].includes(method)
    ? undefined
    : await readBody(event);

  try {
    const response = await $fetch.raw(target, {
      method,
      query: getQuery(event),
      body,
      headers,
      responseType: isAvatarRequest ? 'arrayBuffer' : undefined,
      retry: 0,
      timeout: 30000,
    });

    const contentType = response.headers.get('content-type');
    if (contentType) {
      setResponseHeader(event, 'content-type', contentType);
    }
    setResponseStatus(event, response.status);
    if (isAvatarRequest) {
      return Buffer.from(response._data);
    }
    return response._data;
  } catch (error) {
    if (error?.response) {
      const status = error.response.status || 500;

      setResponseStatus(event, status);
      setResponseHeader(event, 'content-type', 'application/json');
      return normalizeBackendError(status, error.response._data);
    }
    setResponseStatus(event, 502);
    setResponseHeader(event, 'content-type', 'application/json');
    return {
      error: 'api_unavailable',
      status: 502,
      message: error?.message || 'Monitor API is unavailable',
    };
  }
});

function normalizeBackendError(status, data) {
  if (data && typeof data === 'object' && !Array.isArray(data)) {
    return {
      status,
      ...data,
      error: data.error || `backend_${status}`,
    };
  }

  const message = typeof data === 'string' ? data.trim() : '';

  return {
    error: status === 404 ? 'not_found' : `backend_${status}`,
    status,
    message: message.slice(0, 240),
  };
}
