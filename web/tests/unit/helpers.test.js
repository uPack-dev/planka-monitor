import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  awaitRAF,
  awaitTimeout,
  buildFilterParams,
  buildQueryParams,
  buildSrcset,
  buildStrapiFilters,
  cleanPhone,
  cloneObject,
  compressRanges,
  convertIsoStartEndToDateRange,
  convertIsoToDate,
  convertIsoToQuartal,
  escapeNonNumeric,
  formatBytes,
  formatDate,
  formatMinutes,
  formatNumberAbbr,
  formatPrice,
  formatToQuarter,
  getFontPreloadList,
  getFontsPreloadList,
  minMax,
  normalizeMediaPath,
  parseQueryParams,
  preloadImage,
  preloadImages,
  round,
  slugFromString,
  youtubeParser,
} from '../../app/utils/helpers.js';
import { isMockTasksEnabled } from '../../app/utils/mockTasks.mjs';

afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe('helpers', () => {
  it('clamps and clones values', () => {
    expect(minMax(15, 5, 10)).toBe(10);

    const source = { nested: { value: 1 } };
    const cloned = cloneObject(source);

    cloned.nested.value = 2;

    expect(source.nested.value).toBe(1);
  });

  it('wraps timeout and animation frame in promises', async () => {
    vi.useFakeTimers();

    const timeoutPromise = awaitTimeout(25);
    await vi.advanceTimersByTimeAsync(25);

    await expect(timeoutPromise).resolves.toBeUndefined();

    const requestAnimationFrame = vi.fn((callback) => {
      callback();
      return 1;
    });
    vi.stubGlobal('requestAnimationFrame', requestAnimationFrame);

    await expect(awaitRAF()).resolves.toBeUndefined();
    expect(requestAnimationFrame).toHaveBeenCalledTimes(1);
  });

  it('preloads one or many images and resolves load errors too', async () => {
    class ImmediateImage {
      listeners = {};

      set src(value) {
        this.currentSrc = value;
        queueMicrotask(() => this.listeners.load?.());
      }

      addEventListener(type, callback) {
        this.listeners[type] = callback;
      }
    }

    vi.stubGlobal('Image', ImmediateImage);

    await expect(preloadImage('/image.jpg')).resolves.toBeUndefined();
    await expect(preloadImages(['/a.jpg', '/b.jpg'])).resolves.toEqual([
      undefined,
      undefined,
    ]);
  });

  it('builds font preload descriptors', () => {
    expect(
      getFontPreloadList({ path: 'AlegreyaSans-', weights: [400] }, '/base/'),
    ).toEqual([
      {
        rel: 'preload',
        href: '/base/fonts/AlegreyaSans-400.woff2',
        as: 'font',
        type: 'font/woff2',
        crossorigin: true,
      },
    ]);

    expect(
      getFontsPreloadList(
        [
          { path: 'A-', weights: [400] },
          { path: 'B-', weights: [700] },
        ],
        '/',
      ),
    ).toHaveLength(2);
  });

  it('formats numbers, prices, dates, bytes, and minutes', () => {
    expect(round(1.234, 2)).toBe(1.23);
    expect(formatNumberAbbr(4000)).toBe('4K');
    expect(formatNumberAbbr(1250000, 1)).toBe('1.3M');
    expect(formatNumberAbbr(-1500, 1)).toBe('-1.5K');
    expect(formatToQuarter('2025-07-23')).toBe('Q3 2025');
    expect(formatPrice('1002003.45', ' ')).toBe('1 002 003.45');
    expect(formatPrice(null)).toBe('');
    expect(formatDate('2025-07-11T12:57:48.426Z')).toBe('11.07.2025');
    expect(formatBytes(1024)).toBe('1 KB');
    expect(formatMinutes(90)).toBe('1 hour');
  });

  it('builds and parses filter query data', () => {
    const filters = [
      { name: 'rooms', value: [{ value: 'studio' }, { value: '1br' }] },
      {
        name: 'price',
        value: {
          min: { value: 100 },
          max: { value: 300 },
        },
      },
    ];

    expect(buildStrapiFilters(filters)).toEqual({
      rooms: { $in: ['studio', '1br'] },
      price_aed: { $gte: 100, $lte: 300 },
    });

    expect(buildQueryParams(filters)).toEqual({
      rooms: 'studio,1br',
      price_min: 100,
      price_max: 300,
    });

    expect(parseQueryParams({ price_max: '300', rooms: 'studio,1br' })).toEqual(
      [
        { name: 'rooms', value: ['studio', '1br'] },
        {
          name: 'price',
          value: {
            min: { label: null, value: null },
            max: { label: 300, value: 300 },
          },
        },
      ],
    );
  });

  it('handles string, phone, media, srcset, and numeric utilities', () => {
    expect(youtubeParser('https://www.youtube.com/watch?v=dQw4w9WgXcQ')).toBe(
      'dQw4w9WgXcQ',
    );
    expect(slugFromString('Hello World')).toBe('hello_world');
    expect(cleanPhone('+1 234 567')).toBe('+1234567');
    expect(
      buildSrcset(
        {
          url: '/image.jpg',
          width: 800,
          formats: { large: { url: '/image-large.jpg', width: 1600 } },
        },
        'url',
        ['large'],
      ),
    ).toBe('/image.jpg 800w, /image-large.jpg 1600w');
    expect(normalizeMediaPath('./storage/file.jpg')).toBe('/storage/file.jpg');
    expect(compressRanges([3, 2, 1, 5, 5])).toBe('1-3, 5');
    expect(escapeNonNumeric('+1 (234) 567')).toBe('1234567');
  });

  it('converts ISO dates and keeps invalid values intact', () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});

    expect(convertIsoToDate('2025-10-29T09:39:41Z')).toBe('29.10.2025');
    expect(convertIsoToDate('not-a-date')).toBe('not-a-date');
    expect(convertIsoToQuartal('2025-04-15')).toBe('Q2 2025');
    expect(convertIsoToQuartal('not-a-date')).toBe('not-a-date');
    expect(convertIsoStartEndToDateRange('2024-09-07', '2024-09-08')).toBe(
      'Sep 7 — Sep 8, 2024',
    );
    expect(warn).toHaveBeenCalledWith('Invalid date format:', 'not-a-date');
  });

  it('builds URLSearchParams for filter objects', () => {
    const params = buildFilterParams({
      bedrooms: [1, 2],
      price: { min: 100, max: 300 },
      empty: '',
    });

    expect(params.toString()).toBe(
      'filter%5Bbedrooms%5D%5B%5D=1&filter%5Bbedrooms%5D%5B%5D=2&filter%5Bprice_min%5D=100&filter%5Bprice_max%5D=300',
    );
  });

  it('enables mock tasks only from explicit mode values', () => {
    expect(isMockTasksEnabled()).toBe(false);
    expect(isMockTasksEnabled('')).toBe(false);
    expect(isMockTasksEnabled('dev')).toBe(false);
    expect(isMockTasksEnabled('0')).toBe(false);
    expect(isMockTasksEnabled('false')).toBe(false);
    expect(isMockTasksEnabled('1')).toBe(true);
    expect(isMockTasksEnabled(' true ')).toBe(true);
    expect(isMockTasksEnabled('on')).toBe(true);
    expect(isMockTasksEnabled('auto')).toBe(true);
  });
});
