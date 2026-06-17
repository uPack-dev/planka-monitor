import qs from 'qs';
import { API_PREFIX, API_VERSIONS } from '@/constants/variables';

/**
 * Fetch data from backend api
 * @description Use it in store actions
 * @description Use store actions in useAsyncData
 * @description Provide NUXT_PUBLIC_CMS_URL for the backend base URL
 * @param {string} url request url after api prefix
 * @param {object} [options] request options (query, body, headers etc.)
 * @param {string} [apiVersion] api url version prefix e.g. 'v1'
 * @returns {Promise<any>} request promise
 * @example
 * useRequest('/route').then((data) => console.log(data));
 * @see https://nuxt.com/docs/api/utils/dollarfetch
 * @see https://nuxt.com/docs/api/composables/use-async-data
 * @see https://pinia.vuejs.org/ssr/nuxt.html
 */
export const useRequest = (url, options = {}, apiVersion = API_VERSIONS.v2) => {
  const {
    app: { baseURL },
    public: { cmsUrl },
  } = useRuntimeConfig();

  const appBaseUrl = baseURL === '/' ? '' : baseURL;
  const baseUrl = (cmsUrl || appBaseUrl || '').replace(/\/+$/, '');
  const requestOptions = { ...options };

  if (requestOptions.query) {
    const queryString = qs.stringify(requestOptions.query, {
      encode: false,
      arrayFormat: 'brackets',
      skipNulls: true,
    });

    if (queryString) {
      url = `${url}?${queryString}`;
    }

    requestOptions.query = undefined;
  }

  const mergedOptions = {
    retry: 1,
    retryDelay: 500,
    timeout: 30000,
    ...requestOptions,
    onRequestError({ error }) {
      if (import.meta.client && error.name !== 'AbortError') {
        console.error('Request error:', error);
      }
    },
    onResponseError({ response }) {
      if (import.meta.client) {
        console.error('Response error:', response.status, response.statusText);
      }
    },
  };

  let requestUrl = baseUrl;

  requestUrl += `/${API_PREFIX}`;

  requestUrl += `/${apiVersion}`;

  if (url.startsWith('/')) {
    requestUrl += url;
  } else {
    requestUrl += `/${url}`;
  }

  return $fetch.raw(requestUrl, mergedOptions);
};
