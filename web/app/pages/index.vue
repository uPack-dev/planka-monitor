<template>
  <div class="pages-index">
    <AFade>
      <MiniStatePage
        v-if="isBooting"
        key="loading"
        title="Завантажуємо mini-app"
        text="Перевіряємо Telegram-сесію та ваш Workspace."
      />

      <MiniStatePage
        v-else-if="currentStep === 'auth-error'"
        key="auth"
        mark="!"
        title="Відкрийте з Telegram"
        text="Mini-app працює тільки з кнопки бота, щоб ми могли безпечно перевірити ваш Telegram-профіль."
        muted
      />

      <MiniStatePage
        v-else-if="currentStep === 'service-error'"
        key="service-error"
        mark="!"
        title="Не вдалося увійти"
        text="Telegram-сесія є, але mini-app не отримав відповідь від backend. Спробуйте відкрити ще раз або перевірте MONITOR_API_URL."
        :error="miniApp.errorCode ? serviceErrorLabel : ''"
        muted
      />

      <MiniStatePage
        v-else-if="currentStep === 'needs-start'"
        key="needs-start"
        title="Спершу напишіть /start"
        text="Боту потрібен ваш приватний chat_id. Поверніться в чат з ботом, надішліть /start і відкрийте mini-app ще раз."
      />

      <MiniAppShell
        v-else-if="isOnboardingStep"
        key="onboarding"
        theme="dark"
        variant="onboarding"
      >
        <AFade>
          <MiniWelcomePage
            v-if="currentStep === 'welcome'"
            key="welcome"
            @skip="openBindStep"
            @start="openBindStep"
          />

          <MiniBindPage
            v-else-if="currentStep === 'bind'"
            key="bind"
            :suggested-workspace="suggestedWorkspace"
            :avatar-source="avatarSource"
            :display-name="displayName"
            :telegram-username="telegramUsername"
            :workspace-input="workspaceInput"
            :selected-workspace-mode="selectedWorkspaceMode"
            :binding-error="bindingError"
            :is-loading="miniApp.isLoading"
            @activate-manual-workspace="activateManualWorkspace"
            @select-manual-workspace="selectManualWorkspace"
            @select-suggested-workspace="selectSuggestedWorkspace"
            @update:workspace-input="updateWorkspaceInput"
            @submit="confirmBinding"
          />

          <MiniNotificationsPage
            v-else-if="currentStep === 'notifications'"
            key="notifications"
            :preference-cards="preferenceCards"
            :preferences="draftPreferences"
            :error-code="miniApp.errorCode"
            :preference-error="preferenceError"
            :is-loading="miniApp.isLoading"
            @update-preference="setDraftPreference"
            @finish="finishNotifications"
          />
        </AFade>
      </MiniAppShell>

      <MiniEmptyWorkspacePage
        v-else-if="currentStep === 'no-workspace'"
        key="no-workspace"
        @bind="openBindStep"
      />

      <section v-else key="app" class="monitor-app">
        <MiniTopBar
          :eyebrow="topBarEyebrow"
          :title="topBarTitle"
          :model-value="topBarModel"
          :tabs="topBarTabs"
          :show-search="activeTab !== 'profile'"
          @update:model-value="updateTopBarModel"
          @search="toggleSearch"
          @restart-onboarding="handleRestartOnboarding"
        />

        <main class="monitor-app__content">
          <Transition name="search-panel">
            <form
              v-if="searchOpen && activeTab !== 'profile'"
              class="search-panel"
              @submit.prevent="applySearch"
            >
              <input
                v-model="searchDraft"
                type="search"
                :placeholder="searchPlaceholder"
                enterkeyhint="search"
              />
              <button type="submit">Знайти</button>
            </form>
          </Transition>

          <Transition name="app-error">
            <p v-if="appError" class="app-error">{{ appError }}</p>
          </Transition>

          <Transition name="mini-tab" mode="out-in">
            <MiniTasksPage
              v-if="activeTab === 'tasks'"
              key="tasks"
              :task-view="taskView"
              :current-weekday="currentWeekday"
              :selected-date="selectedDate"
              :current-date-label="currentDateLabel"
              :greeting-label="greetingLabel"
              :task-stats="taskStats"
              :calendar-picker-mode="calendarPickerMode"
              :calendar-month-label="calendarMonthLabel"
              :calendar-year-label="calendarYearLabel"
              :calendar-month-options="calendarMonthOptions"
              :calendar-view-date-object="calendarViewDateObject"
              :calendar-year-range-label="calendarYearRangeLabel"
              :calendar-year-options="calendarYearOptions"
              :calendar-day-names="calendarDayNames"
              :month-calendar-days="monthCalendarDays"
              :selected-date-label="selectedDateLabel"
              :is-loading="miniApp.isLoading"
              :visible-task-items="visibleTaskItems"
              :task-groups="taskGroups"
              @change-calendar-month="changeCalendarMonth"
              @toggle-calendar-picker="toggleCalendarPicker"
              @close-calendar-picker="closeCalendarPicker"
              @select-calendar-month="selectCalendarMonth"
              @change-calendar-year-page="changeCalendarYearPage"
              @select-calendar-year="selectCalendarYear"
              @select-date="selectDate"
              @open-card="openCard"
              @complete="completeTimelineItem"
              @menu="openTaskMenu"
            />

            <MiniFeedPage
              v-else-if="activeTab === 'feed'"
              key="feed"
              :items="miniApp.feedItems"
              :is-loading="miniApp.isLoading"
              :feed-title="feedTitle"
              :relative-date="relativeDate"
              @refresh="loadFeed"
              @open="openFeedItem"
              @menu="openFeedMenu"
            />

            <MiniProfilePage
              v-else-if="activeTab === 'profile'"
              key="profile"
              :avatar-source="avatarSource"
              :display-name="displayName"
              :telegram-username="telegramUsername"
              :workspace-label="workspaceLabel"
              :workspace-url="miniApp.me?.workspaceUrl || ''"
              :task-stats="taskStats"
              :completed-today-count="miniApp.timelineStats.completed"
              :completed-month-count="miniApp.timelineStats.completed"
              :due-this-month-count="dueThisMonthCount"
              :unread-feed-count="unreadFeedCount"
              :notifications-enabled="Boolean(miniApp.me?.notificationsEnabled)"
              :enabled-preference-count="enabledPreferenceCount"
              :is-admin="Boolean(miniApp.me?.isAdmin)"
              :master-notifications="masterNotifications"
              :preference-cards="preferenceCards"
              :preferences="draftPreferences"
              :error-code="miniApp.errorCode"
              :preference-error="preferenceError"
              @toggle-master="toggleMasterNotifications"
              @update-preference="saveProfilePreference"
            />

            <MiniAdminPage
              v-else
              key="admin"
              :is-admin="Boolean(miniApp.me?.isAdmin)"
              :is-loading="miniApp.isLoading"
              :items="miniApp.adminSubscribers"
              :stats="adminStats"
              @reset="resetAdminFilters"
              @block="setSubscriberBlocked"
              @admin="setSubscriberAdmin"
            />
          </Transition>
        </main>

        <MiniTaskActionSheet
          :item="taskActionItem"
          @close="taskActionItem = null"
          @complete="completeTaskActionItem"
          @comment="commentTaskActionItem"
          @mute="muteTaskActionItem"
        />

        <MiniFeedActionSheet
          :item="feedActionItem"
          @close="feedActionItem = null"
          @mute="muteFeedActionItem"
        />

        <MiniBottomNav
          v-model:active="activeTab"
          :is-admin="Boolean(miniApp.me?.isAdmin)"
        />

        <MiniCardSheet
          :card="miniApp.activeCard"
          :workspace-url="activeCardWorkspaceUrl"
          :current-username="currentWorkspaceUsername"
          :composer-request-key="cardComposerRequestKey"
          @close="miniApp.activeCard = null"
          @complete-card="completeCard"
          @complete-task="completeTask"
          @comment="addComment"
          @navigate="navigateFromCardSheet"
        />
      </section>
    </AFade>
  </div>
</template>

<script setup>
import MiniAdminPage from '@/components/mini/screens/AdminPage.vue';
import MiniBindPage from '@/components/mini/screens/BindPage.vue';
import MiniEmptyWorkspacePage from '@/components/mini/screens/EmptyWorkspacePage.vue';
import MiniFeedActionSheet from '@/components/mini/FeedActionSheet.vue';
import MiniFeedPage from '@/components/mini/screens/FeedPage.vue';
import MiniNotificationsPage from '@/components/mini/screens/NotificationsPage.vue';
import MiniProfilePage from '@/components/mini/screens/ProfilePage.vue';
import MiniStatePage from '@/components/mini/screens/StatePage.vue';
import MiniTasksPage from '@/components/mini/screens/TasksPage.vue';
import MiniWelcomePage from '@/components/mini/screens/WelcomePage.vue';
import { useMiniAppErrors } from '@/composables/mini/useMiniAppErrors';
import { useMiniAppNavigation } from '@/composables/mini/useMiniAppNavigation';
import { useMiniAppSession } from '@/composables/mini/useMiniAppSession';
import { useMiniTaskTimeline } from '@/composables/mini/useMiniTaskTimeline';
import { useMiniAppStore } from '@/stores/mini-app';

const miniApp = useMiniAppStore();
const telegram = useTelegramWebApp();
const devAuth = useMonitorDevAuth();
const route = useRoute();
const runtimeConfig = useRuntimeConfig();

const onboardingSteps = new Set(['welcome', 'bind', 'notifications']);
const currentStep = shallowRef('loading');
const isOnboardingStep = computed(() => onboardingSteps.has(currentStep.value));
const activeTab = shallowRef('tasks');
const { appError, withAppError } = useMiniAppErrors(miniApp);
const timeline = {};

function loadTasks() {
  return timeline.loadTasks();
}

function clearTaskActionItem() {
  timeline.clearTaskActionItem?.();
}

const navigation = useMiniAppNavigation({
  activeTab,
  clearTaskActionItem,
  currentStep,
  loadTasks,
  miniApp,
  withAppError,
});
const {
  adminStats,
  applySearch,
  feedTitle,
  loadFeed,
  relativeDate,
  resetAdminFilters,
  searchDraft,
  searchOpen,
  searchPlaceholder,
  searchQuery,
  setSubscriberAdmin,
  setSubscriberBlocked,
  taskView,
  toggleSearch,
  topBarEyebrow,
  topBarModel,
  topBarTabs,
  topBarTitle,
  updateTopBarModel,
} = navigation;

const session = useMiniAppSession({
  activeTab,
  currentStep,
  devAuth,
  loadTasks,
  miniApp,
  telegram,
});
const {
  activateManualWorkspace,
  avatarSource,
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
} = session;

Object.assign(
  timeline,
  useMiniTaskTimeline({
    activeTab,
    displayName,
    miniApp,
    route,
    runtimeConfig,
    searchQuery,
    taskView,
    withAppError,
  }),
);
const {
  activeCardWorkspaceUrl,
  addComment,
  calendarDayNames,
  calendarMonthLabel,
  calendarMonthOptions,
  calendarPickerMode,
  calendarYearLabel,
  calendarYearOptions,
  calendarYearRangeLabel,
  calendarViewDateObject,
  cardComposerRequestKey,
  changeCalendarMonth,
  changeCalendarYearPage,
  closeCalendarPicker,
  commentTaskActionItem,
  completeCard,
  completeTask,
  completeTaskActionItem,
  completeTimelineItem,
  currentDateLabel,
  currentWeekday,
  dueThisMonthCount,
  feedActionItem,
  greetingLabel,
  monthCalendarDays,
  muteFeedActionItem,
  muteTaskActionItem,
  navigateFromCardSheet,
  openCard,
  openFeedMenu,
  openFeedItem,
  openTaskMenu,
  selectCalendarMonth,
  selectCalendarYear,
  selectDate,
  selectedDate,
  selectedDateLabel,
  taskActionItem,
  taskGroups,
  taskStats,
  toggleCalendarPicker,
  visibleTaskItems,
} = timeline;

const unreadFeedCount = computed(
  () => (miniApp.feedItems || []).filter((item) => !item.isRead).length,
);

async function handleRestartOnboarding() {
  await withAppError(
    'Не вдалося перезапустити налаштування',
    restartOnboarding,
  );
}
</script>

<style scoped lang="scss">
.pages-index {
  width: 100%;
  min-height: 100vh;
  min-height: 100dvh;
  min-height: var(--tg-viewport-height, 100dvh);
  background-color: $annonce-color-navy;
}

.monitor-app {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100vh;
  height: 100dvh;
  height: var(--tg-viewport-height, 100dvh);
  overflow: hidden;
  font-family: Inter, sans-serif;
  color: $annonce-color-ink;
  background-color: $annonce-color-cream;

  &__content {
    box-sizing: border-box;
    display: flex;
    flex: 1 1 auto;
    flex-direction: column;
    gap: 14px;
    width: 100%;
    min-height: 0;
    padding: 16px 16px calc(var(--tg-safe-bottom, 0px) + 110px);
    overflow-y: auto;
    scroll-behavior: smooth;
    overscroll-behavior: contain;
    scrollbar-width: none;
    -webkit-overflow-scrolling: touch;

    &::-webkit-scrollbar {
      display: none;
    }

    @include media-breakpoint-up(sm) {
      max-width: 1040px;
      padding-right: 24px;
      padding-left: 24px;
      margin: 0 auto;
    }
  }
}

.search-panel {
  display: grid;
  flex-shrink: 0;
  grid-template-columns: 1fr auto;
  gap: 8px;
  padding: 10px;
  background-color: #fffdfa;
  border: 1.5px solid #d7e4ed;
  border-radius: 12px;
  box-shadow: 0 8px 20px rgba(#0b304b, 0.06);
  transition:
    border-color $time-normal $ease-out,
    box-shadow $time-normal $ease-out,
    transform $time-normal $ease-out;

  &:focus-within {
    border-color: rgba($annonce-color-blue, 0.38);
    box-shadow: 0 12px 28px rgba(#0b304b, 0.12);
    transform: translate3d(0, -1px, 0);
  }

  input {
    min-width: 0;
    font-size: 14px;
    color: $annonce-color-ink;
  }

  button {
    min-height: 36px;
    padding: 0 14px;
    font-size: 13px;
    font-weight: 800;
    color: $annonce-color-text;
    background-color: $annonce-color-blue;
    border-radius: 10px;
    transition:
      background-color $time-fast $ease-out,
      transform $time-fast $ease-out;

    &:active {
      transform: scale(0.98);
    }
  }
}

.app-error {
  flex-shrink: 0;
  padding: 10px 12px;
  font-size: 13px;
  line-height: 1.35;
  color: #8f243e;
  background-color: #fff1f1;
  border: 1px solid #ffd3d3;
  border-radius: 12px;
}

.search-panel-enter-active,
.search-panel-leave-active,
.app-error-enter-active,
.app-error-leave-active {
  transition:
    opacity $time-normal $ease-out,
    transform $time-normal $ease-out,
    filter $time-normal $ease-out;
}

.search-panel-enter-from,
.search-panel-leave-to,
.app-error-enter-from,
.app-error-leave-to {
  opacity: 0;
  filter: blur(3px);
  transform: translate3d(0, -4px, 0) scale(0.98);
}

@media (prefers-reduced-motion: reduce) {
  .monitor-app__content {
    scroll-behavior: auto;
  }

  .search-panel,
  .search-panel button,
  .search-panel-enter-active,
  .search-panel-leave-active,
  .app-error-enter-active,
  .app-error-leave-active {
    transition-duration: 1ms;
  }

  .search-panel-enter-from,
  .search-panel-leave-to,
  .app-error-enter-from,
  .app-error-leave-to {
    filter: none;
    transform: none;
  }
}
</style>
