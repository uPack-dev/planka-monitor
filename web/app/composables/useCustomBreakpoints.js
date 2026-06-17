import { BREAKPOINTS } from '@/constants/breakpoints';

/**
 * Get css bootstrap-like breakpoints in js
 * @returns {{greaterOrEqual: function, smallerOrEqual: function, greater: function, smaller: function, between: function, isGreater: function, isGreaterOrEqual: function, isSmaller: function, isSmallerOrEqual: function, isInBetween: function}} breakpoint methods object
 * @example
 * const breakpoint = useCustomBreakpoints();
 * const isSmBreakpoint = breakpoint.smaller('sm');
 * @see https://vueuse.org/useBreakpoints
 */
export const useCustomBreakpoints = () => {
  return useBreakpoints(BREAKPOINTS);
};
