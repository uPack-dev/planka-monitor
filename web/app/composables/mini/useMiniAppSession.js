import { DEFAULT_PREFERENCES } from '@/stores/mini-app';

const TELEGRAM_BOOT_TIMEOUT_MS = 5000;
const API_BOOT_TIMEOUT_MS = 10000;

const preferenceCards = [
  {
    key: 'assignments',
    icon: 'user',
    title: 'Призначення',
    text: 'Коли вас додали до картки/задачі',
  },
  {
    key: 'comments',
    icon: 'message',
    title: 'Коментарі',
    text: 'Коментар у картці, де ви учасник',
  },
  {
    key: 'changes',
    icon: 'pencil',
    title: 'Зміни',
    text: 'Оновлення опису або дедлайну',
  },
  {
    key: 'done',
    icon: 'check',
    title: 'Виконано',
    text: 'Хтось закрив вашу задачу',
  },
];

export function useMiniAppSession({
  activeTab,
  currentStep,
  devAuth,
  loadTasks,
  miniApp,
  telegram,
}) {
  const workspaceInput = shallowRef('');
  const selectedWorkspaceMode = shallowRef('manual');
  const draftPreferences = reactive({ ...DEFAULT_PREFERENCES });
  const masterNotifications = shallowRef(false);

  const isBooting = computed(() => currentStep.value === 'loading');
  const telegramUsername = computed(() => miniApp.me?.telegramUsername || '');
  const suggestedWorkspace = computed(
    () => miniApp.me?.workspaceUsername || telegramUsername.value,
  );
  const workspaceLabel = computed(
    () => miniApp.me?.workspaceUsername || 'не привʼязано',
  );
  const currentWorkspaceUsername = computed(
    () => miniApp.me?.workspaceUsername || telegramUsername.value,
  );
  const workspaceDisplayName = computed(
    () => miniApp.me?.workspaceDisplayName || '',
  );
  const displayName = computed(
    () =>
      workspaceDisplayName.value ||
      currentWorkspaceUsername.value ||
      telegramUsername.value ||
      'Workspace user',
  );
  const avatarSource = computed(() => miniApp.me?.workspaceAvatarUrl || '');
  const enabledPreferenceCount = computed(
    () => Object.values(draftPreferences).filter(Boolean).length,
  );
  const bindingError = computed(() => {
    switch (miniApp.errorCode) {
      case 'not_found':
        return 'Такого Workspace username не знайдено.';
      case 'already_bound':
        return 'Цей Workspace username вже привʼязаний до іншого Telegram.';
      case 'needs_start':
        return 'Спершу напишіть /start боту в приватному чаті.';
      case 'username_required':
        return 'У Telegram має бути встановлений username.';
      default:
        return miniApp.errorCode
          ? 'Не вдалося зберегти привʼязку. Спробуйте ще раз.'
          : '';
    }
  });
  const preferenceError = computed(() => {
    switch (miniApp.errorCode) {
      case 'needs_start':
        return 'Спершу напишіть /start боту в приватному чаті.';
      case 'unauthorized':
        return 'Telegram-сесія застаріла. Відкрийте mini-app ще раз із бота.';
      default:
        return 'Не вдалося зберегти налаштування. Спробуйте ще раз.';
    }
  });
  const serviceErrorLabel = computed(() => {
    const status = miniApp.errorStatus ? ` · HTTP ${miniApp.errorStatus}` : '';
    return `${miniApp.errorCode}${status}`;
  });

  onMounted(bootMiniApp);

  async function bootMiniApp() {
    devAuth.syncFromURL();

    if (devAuth.hasToken.value) {
      telegram.init();
    } else {
      try {
        await withTimeout(
          telegram.waitForInitData(),
          TELEGRAM_BOOT_TIMEOUT_MS,
          'telegram_init_timeout',
        );
      } catch {
        // The auth-error state below will explain missing Telegram init data.
      }
    }
    telegram.ready();
    telegram.expand();

    if (!telegram.initData.value && !devAuth.hasToken.value) {
      currentStep.value = 'auth-error';
      return;
    }

    try {
      await withTimeout(
        miniApp.fetchMe(),
        API_BOOT_TIMEOUT_MS,
        'api_boot_timeout',
      );
      hydrateLocalState();

      if (miniApp.me?.needsStart) {
        currentStep.value = 'needs-start';
      } else if (!miniApp.me?.onboardingCompleted) {
        currentStep.value = 'welcome';
      } else if (miniApp.hasWorkspace) {
        await enterApp();
      } else {
        currentStep.value = 'no-workspace';
      }
    } catch {
      if (!miniApp.errorCode) {
        miniApp.errorCode = 'boot_timeout';
        miniApp.errorStatus = 0;
      }
      currentStep.value =
        miniApp.errorCode === 'unauthorized' ? 'auth-error' : 'service-error';
    }
  }

  function withTimeout(promise, timeout, errorCode) {
    let timer;
    const timeoutPromise = new Promise((resolve, reject) => {
      timer = setTimeout(() => reject(new Error(errorCode)), timeout);
    });

    return Promise.race([promise, timeoutPromise]).finally(() => {
      clearTimeout(timer);
    });
  }

  function hydrateLocalState() {
    workspaceInput.value = suggestedWorkspace.value;
    selectedWorkspaceMode.value = suggestedWorkspace.value
      ? 'suggested'
      : 'manual';
    Object.assign(draftPreferences, miniApp.preferences);
    masterNotifications.value = Boolean(miniApp.me?.notificationsEnabled);
  }

  async function enterApp() {
    activeTab.value = 'tasks';
    currentStep.value = 'app';
    await loadTasks();
  }

  function openBindStep() {
    miniApp.clearError();
    workspaceInput.value = suggestedWorkspace.value;
    selectedWorkspaceMode.value = suggestedWorkspace.value
      ? 'suggested'
      : 'manual';
    currentStep.value = 'bind';
  }

  function activateManualWorkspace() {
    miniApp.clearError();
    selectedWorkspaceMode.value = 'manual';
  }

  function selectManualWorkspace() {
    activateManualWorkspace();
    workspaceInput.value = '';
  }

  function selectSuggestedWorkspace() {
    if (!suggestedWorkspace.value) return;

    miniApp.clearError();
    selectedWorkspaceMode.value = 'suggested';
    workspaceInput.value = suggestedWorkspace.value;
  }

  function updateWorkspaceInput(value) {
    selectedWorkspaceMode.value = 'manual';
    workspaceInput.value = value;
  }

  async function confirmBinding() {
    miniApp.clearError();
    const username =
      selectedWorkspaceMode.value === 'suggested'
        ? suggestedWorkspace.value.trim()
        : workspaceInput.value.trim();
    if (!username) return;

    try {
      await miniApp.bindWorkspace(username);
      hydrateLocalState();
      currentStep.value = 'notifications';
    } catch {
      // Error text is derived from store.errorCode.
    }
  }

  async function finishNotifications() {
    miniApp.clearError();
    try {
      await miniApp.savePreferences({ ...draftPreferences });
      await miniApp.setMasterNotifications(true);
      await miniApp.setOnboardingCompleted(true);
      hydrateLocalState();
      await enterApp();
    } catch {
      // Error text is derived from store.errorCode.
    }
  }

  function setDraftPreference(key, value) {
    draftPreferences[key] = value;
  }

  async function saveProfilePreference(key, value) {
    draftPreferences[key] = value;
    miniApp.clearError();
    try {
      await miniApp.savePreferences({ ...draftPreferences });
    } catch {
      draftPreferences[key] = !value;
    }
  }

  async function restartOnboarding() {
    miniApp.clearError();
    await miniApp.setOnboardingCompleted(false);
    hydrateLocalState();
    currentStep.value = 'welcome';
  }

  async function toggleMasterNotifications(value) {
    masterNotifications.value = value;
    miniApp.clearError();
    try {
      await miniApp.setMasterNotifications(value);
    } catch {
      masterNotifications.value = !value;
    }
  }

  return {
    avatarSource,
    activateManualWorkspace,
    bindingError,
    confirmBinding,
    currentWorkspaceUsername,
    displayName,
    draftPreferences,
    enabledPreferenceCount,
    finishNotifications,
    isBooting,
    masterNotifications,
    openBindStep,
    preferenceCards,
    preferenceError,
    restartOnboarding,
    saveProfilePreference,
    serviceErrorLabel,
    setDraftPreference,
    selectManualWorkspace,
    selectSuggestedWorkspace,
    selectedWorkspaceMode,
    suggestedWorkspace,
    telegramUsername,
    toggleMasterNotifications,
    updateWorkspaceInput,
    workspaceInput,
    workspaceLabel,
  };
}
