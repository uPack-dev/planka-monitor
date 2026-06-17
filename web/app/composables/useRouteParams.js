/**
 * Extracts the slug (project name) from a URL path
 * @param {string} url - The URL path, e.g., "/projects/gianfranco-ferre-residences"
 * @returns {string} The extracted slug, e.g., "gianfranco-ferre-residences"
 */
import { useGlobalStore } from '@/stores/global';

export const extractSlugFromUrl = (url) => {
  if (!url) return '';

  // Remove leading slash if present
  const cleanUrl = url.startsWith('/') ? url.substring(1) : url;

  // Split the URL by '/'
  const segments = cleanUrl.split('/');

  // For URLs like "/projects/gianfranco-ferre-residences"
  // The slug is the second segment
  if (segments.length >= 2) {
    return segments[1];
  }

  return '';
};

export const useRouteParams = (languages) => {
  const globalStore = useGlobalStore();
  const route = useRoute();
  const $languages = globalStore.languages;
  const locales =
    Array.isArray($languages) && $languages.length
      ? $languages
      : Array.isArray(languages)
        ? languages
        : [];

  const localesCodes = locales.map((l) => l.code);
  const defaultLocale =
    locales.find((l) => l.isDefault)?.code || localesCodes[0] || '';
  const isValidLang = (lang) => localesCodes.includes(lang);

  function calcParams() {
    let params = {
      lang: defaultLocale,
      page: '',
      slug: '',
    };
    const { all: segments } = route.params;
    if (segments && segments.length) {
      if (isValidLang(segments[0])) {
        params = {
          lang: segments[0],
          page: segments[1] || '',
          slug: segments[2] || '',
        };
      } else {
        params = {
          lang: defaultLocale,
          page: segments[0] || '',
          slug: segments[1] || '',
        };
      }
    }
    return params;
  }

  return {
    ...calcParams(),
    calcParams,
  };
};
