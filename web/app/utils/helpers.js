// common

/**
 * Get value with boundaries
 * @param {number} value desired value
 * @param {number} min min boundary
 * @param {number} max max boundary
 * @returns {number} calculated value
 * @example
 * minMax(15, 5, 10); // returns 10 (max boundary)
 */
export function minMax(value, min, max) {
  return Math.min(Math.max(value, min), max);
}

/**
 * Simple object clone
 * @param {object} object object to clone
 * @returns {any} cloned object
 * @example
 * const clonedObject = cloneObject(originalObject);
 */
export function cloneObject(object) {
  return JSON.parse(JSON.stringify(object));
}

// promise-based utils
/**
 * Await timeout
 * @param {number} [duration] timeout duration
 * @returns {Promise<void>} timeout promise
 * @example
 * await awaitTimeout(100);
 */
export function awaitTimeout(duration = 0) {
  return new Promise((resolve) => setTimeout(resolve, duration));
}

/**
 * Await request animation frame
 * @returns {Promise<void>} request animation frame promise
 * @example
 * await awaitRAF();
 */
export function awaitRAF() {
  return new Promise((resolve) => requestAnimationFrame(resolve));
}

// preload images
/**
 * Preload image
 * @param {string} url image url
 * @returns {Promise<void>} upload promise
 */
export function preloadImage(url) {
  return new Promise((resolve) => {
    const image = new Image();
    image.src = url;

    image.addEventListener('load', resolve);
    image.addEventListener('error', resolve);
  });
}

/**
 * Preload images
 * @param {string[]} imageUrls image urls
 * @returns {Promise<Awaited<void>[]>} upload promises
 */
export function preloadImages(imageUrls = []) {
  const preloads = Array.from(imageUrls).map(preloadImage);

  return Promise.all(preloads);
}

// preload fonts
/**
 * Get font preload list
 * @param {string} path font path
 * @param {number[]} weights font weights
 * @param {string} baseUrl site base url
 * @returns {{rel: string, href: string, as: string, type: string, crossorigin: true}[]} font preload list
 */
export function getFontPreloadList({ path, weights }, baseUrl = '/') {
  return weights.map((weight) => ({
    rel: 'preload',
    href: `${baseUrl}fonts/${path}${weight}.woff2`,
    as: 'font',
    type: 'font/woff2',
    crossorigin: true,
  }));
}

/**
 * Get fonts preload list
 * @param {object[]} fontsList font params list
 * @param {string} fontsList.path font path
 * @param {number[]} fontsList.weights font weights
 * @param {string} [baseUrl] site base url
 * @returns {{rel: string, href: string, as: string, type: string, crossorigin: true}[]} fonts preload list
 */
export function getFontsPreloadList(fontsList, baseUrl = '/') {
  return fontsList.reduce(
    (result, { path, weights }) => [
      ...result,
      ...getFontPreloadList({ path, weights }, baseUrl),
    ],
    [],
  );
}

export function round(value, digits = 0) {
  const pow = Math.pow(10, digits);
  return Math.round(value * pow) / pow;
}

/**
 * Преобразует число в сокращённый строковый формат:
 * - >=1_000_000_000 → X.YB
 * - >=1_000_000     → X.YM
 * - >=1_000         → X.YK
 * - иначе           → как есть
 *
 * @param {number|string} input — число или строка с цифрами (можно с пробелами)
 * @param {number} [decimals=2] — количество знаков после запятой
 * @returns {string} Например: "1.3M", "850", "4K"
 */
export function formatNumberAbbr(input, decimals = 2) {
  // Убираем пробелы и приводим к числу
  const num =
    typeof input === 'string' ? parseFloat(input.replace(/\s+/g, '')) : input;

  if (!Number.isFinite(num)) {
    return input;
  }

  const abs = Math.abs(num);
  const sign = num < 0 ? '-' : '';

  // Вспомогательная функция форматирования
  function fmt(divisor, suffix) {
    const clean = (Math.abs(num) / divisor)
      .toFixed(decimals)
      .replace(/\.?0+$/, '');
    return sign + clean + suffix;
  }

  if (abs >= 1e9) {
    return fmt(1e9, 'B');
  }
  if (abs >= 1e6) {
    return fmt(1e6, 'M');
  }
  if (abs >= 1e3) {
    return fmt(1e3, 'K');
  }
  // Меньше тысячи — просто возвращаем число без изменений
  return String(num);
}

/**
 * Переводит дату из формата YYYY-MM-DD в формат Q{n} YYYY
 * @param {string} isoDate — строка даты, например "2025-07-23"
 * @returns {string} — строка вида "Q3 2025"
 */
export function formatToQuarter(isoDate = '') {
  // Разбиваем по дефисам и получаем год и месяц
  const [year, month] = isoDate.split('-').map(Number);
  // Квартал: (0–2)→1, (3–5)→2, (6–8)→3, (9–11)→4
  const quarter = Math.floor((month - 1) / 3) + 1;
  return `Q${quarter} ${year}`;
}

/**
 * Форматирует число или строку с цифрами, вставляя пробелы как разделители тысяч.
 * Примеры:
 *   formatPrice(1200000)      // "1 200 000"
 *   formatPrice("1200000")    // "1 200 000"
 *   formatPrice("1002003.45") // "1 002 003.45"
 *
 * @param {number|string} value — исходное значение (число или строка с цифрами и, возможно, пробелами)
 * @param {string} [thousandSeparator=','] — строка, используемая как разделитель разрядов
 * @returns {string} — строка с разделением пробелами по тысячам
 */
export function formatPrice(value, thousandSeparator = ',') {
  if (value == null || value === '') {
    return '';
  }

  // Приводим к строке без пробелов
  let str =
    typeof value === 'number'
      ? String(value)
      : value.toString().replace(/\s+/g, '');

  // Сохраняем знак, если есть
  const sign = str.startsWith('-') ? '-' : '';
  if (sign) str = str.slice(1);

  // Разделяем целую и дробную части
  const [intPart, fracPart] = str.split('.');

  // Вставляем разделитель перед каждой группой из трёх цифр с конца
  const intFormatted = intPart.replace(
    /\B(?=(\d{3})+(?!\d))/g,
    thousandSeparator,
  );

  // Собираем обратно, добавляем дробную часть, если она была
  return sign + intFormatted + (fracPart !== undefined ? '.' + fracPart : '');
}

/**
 * Переводит дату из формата ISO в формат 2025.02.01
 * @param {string} isoDate — строка даты, например "2025-07-11T12:57:48.426Z"
 * @returns {string} — строка вида "Q3 2025"
 */
export const formatDate = (dateString) => {
  if (!dateString) return '';

  const date = new Date(dateString);
  const day = date.getDate().toString().padStart(2, '0');
  const month = (date.getMonth() + 1).toString().padStart(2, '0');
  const year = date.getFullYear();

  return `${day}.${month}.${year}`;
};

export const formatBytes = (a, b) => {
  if (0 === a) return '0 Bytes';

  const c = 1024,
    d = b || 2,
    e = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'],
    f = Math.floor(Math.log(a) / Math.log(c));

  return parseFloat((a / Math.pow(c, f)).toFixed(d)) + ' ' + e[f];
};

export function buildStrapiFilters(filterArray) {
  const filters = {};

  filterArray.forEach(({ name, value }) => {
    // Пропускаем пустые значения
    if (
      value == null ||
      (Array.isArray(value) && value.length === 0) ||
      (typeof value === 'object' &&
        !Array.isArray(value) &&
        value.min?.value == null &&
        value.max?.value == null)
    ) {
      return;
    }

    // 1) Диапазон цен
    if (
      name === 'price' &&
      typeof value === 'object' &&
      !Array.isArray(value)
    ) {
      filters.price_aed = {};
      if (value.min.value != null)
        filters.price_aed.$gte = Number(value.min?.value || 0);
      if (value.max.value != null)
        filters.price_aed.$lte = Number(value.max?.value || Infinity);
      return;
    }

    const filtersData = Array.isArray(value) ? value : [value];
    const valuesFromFiltersData = filtersData.map((data) => data.value);
    // 2) Вложенный фильтр по slug для связанной коллекции project(s)
    if (name === 'projects' || name === 'project') {
      // Если пришёл просто строкой, приводим к массиву

      // Здесь — любой оператор: $in (несколько), или $eq (одно значение)
      filters[name] = {
        slug: { $in: valuesFromFiltersData },
      };
      return;
    }

    // 3) Обычный фильтр по полю в массиве
    filters[name] = {
      $in: valuesFromFiltersData,
    };
  });

  return filters;
}

export function buildQueryParams(filterArray) {
  const query = {};

  filterArray.forEach(({ name, value }) => {
    if (name === 'price' && typeof value === 'object') {
      if (value.min.value != null) query.price_min = value.min.value;
      if (value.max.value != null) query.price_max = value.max.value;
      return;
    }
    // массив → "a,b,c"; одиночное значение тоже оборачиваем в массив
    const vals = Array.isArray(value) ? value : [value];
    if (vals.length) {
      query[name] = vals.map((data) => data.value).join(',');
    }
  });

  return query;
}

export function parseQueryParams(query) {
  const filters = [];
  let priceMin = null,
    priceMax = null;

  Object.entries(query).forEach(([key, val]) => {
    if (key === 'price_min') {
      priceMin = val;
      return;
    }
    if (key === 'price_max') {
      priceMax = val;
      return;
    }
    // всё остальное: разбить по запятым
    const arr = ('' + val).split(',').filter(Boolean);
    filters.push({ name: key, value: arr });
  });

  if (priceMin != null || priceMax != null) {
    filters.push({
      name: 'price',
      value: {
        min: {
          label: priceMin != null ? Number(priceMin) : null,
          value: priceMin != null ? Number(priceMin) : null,
        },
        max: {
          label: priceMax != null ? Number(priceMax) : null,
          value: priceMax != null ? Number(priceMax) : null,
        },
      },
    });
  }

  return filters;
}

export function youtubeParser(url) {
  const regExp =
    /^.*((youtu.be\/)|(v\/)|(\/u\/\w\/)|(embed\/)|(watch\?))\??v?=?([^#&?]*).*/;
  const match = url.match(regExp);
  return match && match[7].length == 11 ? match[7] : false;
}

export function slugFromString(name) {
  return name.toLowerCase().split(' ').join('_');
}

export const cleanPhone = (phone) => phone.replace(/\s/g, '');

/**
 * Builds a valid srcset string for <img> or <source> tags based on the image object and its formats.
 *
 * @param {object} obj - The image object containing the main url, width, and formats (e.g., large, medium, small).
 * @param {string} [mainKey='url'] - The key for the main image url in the object (usually 'url').
 * @param {string[]|string} [formatKeys=[]] - Array of format keys (or a single key) to include from obj.formats (e.g., ['large', 'medium', 'small']).
 * @returns {string} A srcset string, e.g. "/uploads/image.png 800w, /uploads/large_image.png 1600w, /uploads/medium_image.png 750w".
 *
 * @example
 * const image = {
 *   url: '/uploads/image.png',
 *   width: 800,
 *   formats: {
 *     large: { url: '/uploads/large_image.png', width: 1600 },
 *     medium: { url: '/uploads/medium_image.png', width: 750 },
 *   },
 * };
 *
 * const srcset = buildSrcset(image, 'url', ['large', 'medium']);
 * // srcset: "/uploads/image.png 800w, /uploads/large_image.png 1600w, /uploads/medium_image.png 750w"
 */
export function buildSrcset(obj, mainKey = 'url', formatKeys = []) {
  if (!obj) return '';
  const srcs = [];

  if (obj[mainKey] && obj.width) srcs.push(`${obj[mainKey]} ${obj.width}w`);

  if (Array.isArray(formatKeys)) {
    formatKeys.forEach((key) => {
      if (obj.formats?.[key]?.url && obj.formats?.[key]?.width) {
        srcs.push(`${obj.formats[key].url} ${obj.formats[key].width}w`);
      }
    });
  } else if (
    typeof formatKeys === 'string' &&
    obj.formats?.[formatKeys]?.url &&
    obj.formats?.[formatKeys]?.width
  ) {
    srcs.push(
      `${obj.formats[formatKeys].url} ${obj.formats[formatKeys].width}w`,
    );
  }

  return srcs.join(', ');
}

/**
 * Converts ISO date string to DD.MM.YYYY format
 * @param {string} isoString - ISO date string (e.g., "2025-10-29T09:39:41Z")
 * @returns {string} formatted date string (e.g., "29.10.2025")
 * @example
 * convertIsoToDate("2025-10-29T09:39:41Z"); // returns "29.10.2025"
 * convertIsoToDate("2025-07-11"); // returns "11.07.2025"
 */
export function convertIsoToDate(isoString) {
  try {
    const dateObj = new Date(isoString);

    if (Number.isNaN(dateObj.getTime())) {
      throw new Error('Invalid Date');
    }

    const day = dateObj.getDate().toString().padStart(2, '0');
    const month = (dateObj.getMonth() + 1).toString().padStart(2, '0');
    const year = dateObj.getFullYear();

    return `${day}.${month}.${year}`;
  } catch {
    console.warn('Invalid date format:', isoString);
    return isoString;
  }
}

/**
 * Converts ISO date string to Quarter format (QX YYYY)
 * @param {string} isoString - ISO date string (e.g., "2025-10-29T09:39:41Z")
 * @returns {string} formatted quarter string (e.g., "Q4 2025")
 * @example
 * convertIsoToQuartal("2025-04-15"); // returns "Q2 2025"
 * convertIsoToQuartal("2025-10-29T09:39:41Z"); // returns "Q4 2025"
 */
export function convertIsoToQuartal(isoString) {
  try {
    const dateObj = new Date(isoString);

    // Check if date is valid
    if (isNaN(dateObj.getTime())) {
      throw new Error('Invalid Date');
    }

    const month = dateObj.getMonth(); // 0-11
    const quarter = Math.floor(month / 3) + 1;
    const year = dateObj.getFullYear();

    return `Q${quarter} ${year}`;
  } catch {
    console.warn('Invalid date format:', isoString);
    return isoString;
  }
}

/**
 * Normalizes media path:
 * - Keeps external URLs as is (http(s):// or //)
 * - Keeps data URIs as is
 * - Adds a leading '/' to relative paths (e.g. "storage/file.jpg" → "/storage/file.jpg")
 * - Leaves paths starting with '/' unchanged
 * - Returns empty string for null/undefined/empty inputs
 *
 * @param {string|null|undefined} path
 * @returns {string} normalized path
 */
export function normalizeMediaPath(path) {
  if (!path && path !== '') return ''; // handle null/undefined
  path = String(path).trim();

  if (path === '') return ''; // empty string

  // data URI (data:image/...)
  if (/^data:/i.test(path)) return path;

  // absolute URL (http://, https://) or protocol-relative (//)
  if (/^(https?:)?\/\//i.test(path)) return path;

  // already site-absolute (starts with '/')
  if (path.startsWith('/')) return path;

  // remove leading './' if present
  if (path.startsWith('./')) path = path.slice(2);

  // add leading '/' for relative paths
  return '/' + path;
}

/**
 * Convert minutes to readable format (min/hour)
 * @param {number} minutes number of minutes
 * @returns {string} formatted time string
 * @example
 * formatMinutes(3); // returns "3min"
 * formatMinutes(180); // returns "3hour"
 * formatMinutes(90); // returns "1hour"
 */
export function formatMinutes(minutes) {
  if (!minutes || minutes < 60) {
    return `${minutes} min`;
  }

  const hours = Math.floor(minutes / 60);
  return `${hours} hour`;
}

/**
 * Builds a flat filter object for API requests using `URLSearchParams`.
 *
 * Rules:
 *  • Arrays → converted to repeated parameters:
 *      filter[key][]=1&filter[key][]=2
 *
 *  • Objects → flattened using "key_subKey" and appended as:
 *      filter[key_subKey]=value
 *
 *  • Primitives → appended as:
 *      filter[key]=value
 *
 *  • Skips: null, undefined, empty strings, and empty arrays.
 *
 * @param {Object} filters - The filters object to process.
 * @returns {URLSearchParams} URLSearchParams instance with formatted filters.
 *
 * @example
 * buildFilterParams({
 *   bedrooms: [1, 2],
 *   price: { min: 100, max: 300 },
 * })
 *
 * Result:
 *  filter[bedrooms][]=1
 *  filter[bedrooms][]=2
 *  filter[price_min]=100
 *  filter[price_max]=300
 */
export const buildFilterParams = (filters) => {
  const params = new URLSearchParams();

  for (const [key, value] of Object.entries(filters)) {
    if (
      value == null ||
      value === '' ||
      (Array.isArray(value) && value.length === 0)
    ) {
      continue;
    }

    if (Array.isArray(value)) {
      value.forEach((v) => {
        if (v != null && v !== '') {
          params.append(`filter[${key}][]`, v);
        }
      });
      continue;
    }

    if (typeof value === 'object' && !Array.isArray(value)) {
      for (const [subKey, subValue] of Object.entries(value)) {
        if (subValue != null && subValue !== '')
          params.append(`filter[${key}_${subKey}]`, subValue);
      }
      continue;
    }

    params.append(`filter[${key}]`, value);
  }

  return params;
};

/**
 * Converts ISO date strings to formatted date range
 * @param {string|null} isoStart - Start date in ISO format (e.g., "2025-11-07T00:00:00Z")
 * @param {string|null} isoEnd - End date in ISO format (e.g., "2025-11-29T00:00:00Z")
 * @returns {string} Formatted date range string
 * @example
 * // Single date (only start)
 * convertIsoStartEndToDateRange("2024-08-28T00:00:00Z", null); // returns "Aug 28, 2024"
 *
 * // Single date (only end)
 * convertIsoStartEndToDateRange(null, "2024-08-28T00:00:00Z"); // returns "Aug 28, 2024"
 *
 * // Date range
 * convertIsoStartEndToDateRange("2024-09-07T00:00:00Z", "2024-09-08T00:00:00Z"); // returns "Sep 7 — Sep 8, 2024"
 *
 * // No dates
 * convertIsoStartEndToDateRange(null, null); // returns ""
 */
export function convertIsoStartEndToDateRange(isoStart, isoEnd) {
  const startDate = isoStart ? new Date(isoStart) : null;
  const endDate = isoEnd ? new Date(isoEnd) : null;

  const formatDate = (date) => {
    const month = date.toLocaleDateString('en-US', { month: 'short' });
    const day = date.getDate();
    return `${month} ${day}`;
  };

  if (startDate && !endDate) {
    const formatted = formatDate(startDate);
    const year = startDate.getFullYear();
    return `${formatted}, ${year}`;
  }

  if (!startDate && endDate) {
    const formatted = formatDate(endDate);
    const year = endDate.getFullYear();
    return `${formatted}, ${year}`;
  }

  if (startDate && endDate) {
    const startFormatted = formatDate(startDate);
    const endFormatted = formatDate(endDate);
    const year = endDate.getFullYear();
    return `${startFormatted} — ${endFormatted}, ${year}`;
  }

  return '';
}

/**
 * @description Объединяет последовательные числа массива в диапазоны.
 * Функция автоматически удаляет дубликаты и сортирует массив перед обработкой.
 *
 * @param {number[]} nums - Массив целых чисел для обработки.
 * @returns {string} Строка с диапазонами (например, "1-3, 5, 8-9").
 *
 * @example
 * compressRanges([1, 2, 3, 4]); // returns "1-4"
 * compressRanges([1, 4, 6]);    // returns "1, 4, 6"
 * compressRanges([1, 2, 3, 5]); // returns "1-3, 5"
 */
export function compressRanges(nums) {
  if (!nums || nums.length === 0) return '';

  const sortedNums = [...new Set(nums)].sort((a, b) => a - b);

  const ranges = [];
  let start = sortedNums[0];
  let prev = sortedNums[0];

  for (let i = 1; i < sortedNums.length; i++) {
    const curr = sortedNums[i];

    if (curr === prev + 1) {
      prev = curr;
    } else {
      ranges.push(start === prev ? `${start}` : `${start}-${prev}`);
      start = curr;
      prev = curr;
    }
  }

  ranges.push(start === prev ? `${start}` : `${start}-${prev}`);

  return ranges.join(', ');
}

export function escapeNonNumeric(str) {
  return str.replace(/\D/g, '');
}
