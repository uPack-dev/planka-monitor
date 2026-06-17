<template>
  <div id="app" class="app">
    <NuxtLayout v-slot="{ className }" class="app__layout">
      <div :class="className">
        <NuxtPage :keepalive="false" />
      </div>
    </NuxtLayout>
  </div>
</template>

<script setup>
import { useVfm } from 'vue-final-modal';
import { useFeedbackStore } from '@/stores/feedback';

const route = useRoute();
const { closeAll } = useVfm();
const url = route.fullPath;

if (url.includes('?')) {
  const feedbackStore = useFeedbackStore();
  let utmParams = {};

  const params = new URLSearchParams(url.split('?')[1]);
  const allParams = Object.fromEntries(params.entries());
  const allowedUtmParams = [
    'utm_source',
    'utm_medium',
    'utm_campaign',
    'utm_referrer',
    'utm_content',
    'utm_term',
  ];

  utmParams = allowedUtmParams.reduce((obj, key) => {
    if (allParams[key]) {
      obj[key] = allParams[key];
    }
    return obj;
  }, {});

  feedbackStore.setUtm(utmParams);
}

const {
  public: { siteName },
} = useRuntimeConfig();

useMotionPreference();

useHead({
  titleTemplate: (titleChunk) => {
    return titleChunk ? `${titleChunk} | ${siteName}` : siteName;
  },
});

const stopWatcher = watch(
  route,
  (newRoute, oldRoute) => {
    // Don't close modals if only the hash changed
    if (newRoute.fullPath.split('#')[0] === oldRoute.fullPath.split('#')[0]) {
      return;
    }
    closeAll();
  },
  { deep: true },
);
onBeforeUnmount(() => {
  stopWatcher();
});

defineOgImage(
  'BrandedTheme',
  {
    title: siteName,
    description:
      'Telegram mini-app for Workspace notification preferences and account binding.',
  },
  { extension: 'png' },
);
</script>

<style scoped lang="scss">
.app {
  $parent: &;

  display: flex;
  flex-direction: column;

  &:deep(#{$parent}__layout) {
    flex-grow: 1;
  }

  @include rtl {
    direction: rtl;
  }
}
</style>
