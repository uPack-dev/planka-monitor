import {
  disablePageScroll,
  enablePageScroll,
  updateAllScrollbarWidthAdjustment,
} from '@fluejs/noscroll';

const getPageScrollBarWidth = () => {
  if (!import.meta.client) {
    return 0;
  }

  return window.innerWidth - document.documentElement.clientWidth;
};

/**
 * Scroll lock integration
 * @description Fill gap for fixed elements: [data-noscroll-target-scrollbar-width-adjustment]
 * @description Allow scroll for elements: [data-noscroll-target-scrollable]
 * @returns {{lock: function, unlock: function, fillGaps: function, getScrollBarWidth: function}} scroll lock methods object
 * @example
 * const scrollLock = useScrollLock();
 * scrollLock.lock();
 * @see https://github.com/fluejs/noscroll
 */
export const useScrollLock = () => {
  return {
    lock: disablePageScroll,
    unlock: enablePageScroll,

    fillGaps: updateAllScrollbarWidthAdjustment,
    getScrollBarWidth: getPageScrollBarWidth,
  };
};
