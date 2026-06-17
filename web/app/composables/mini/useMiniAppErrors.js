export function useMiniAppErrors(miniApp) {
  const appError = shallowRef('');

  async function withAppError(errorPrefix, action) {
    appError.value = '';
    try {
      await action();
      return true;
    } catch {
      appError.value = apiErrorLabel(errorPrefix);
      return false;
    }
  }

  function apiErrorLabel(prefix) {
    const status = miniApp.errorStatus ? ` · HTTP ${miniApp.errorStatus}` : '';
    return `${prefix}${status}`;
  }

  return {
    appError,
    withAppError,
  };
}
