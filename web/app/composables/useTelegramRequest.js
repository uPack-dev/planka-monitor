import { API_VERSIONS } from '@/constants/variables';

export const useTelegramRequest = (url, options = {}) => {
  const { initData } = useTelegramWebApp();
  const devAuth = useMonitorDevAuth();

  const headers = {
    ...(options.headers || {}),
  };

  if (initData.value) {
    headers['X-Telegram-Init-Data'] = initData.value;
  }
  const devAuthToken = devAuth.getToken();
  if (devAuthToken) {
    headers['X-Monitor-Dev-Auth'] = devAuthToken;
  }

  return useRequest(
    url,
    {
      ...options,
      headers,
    },
    API_VERSIONS.v1,
  );
};
