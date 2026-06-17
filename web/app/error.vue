<template>
  <div id="app" class="error">
    <NuxtLayout v-slot="{ className }" class="error__layout">
      <div v-if="isDev" :class="className">
        <div class="container error__container">
          <p class="h2-r-o error__message">{{ error.statusCode }}</p>

          <p class="h3-r-o error__message">{{ error.message }}</p>

          <UiButton
            class="error__button"
            aria-label="Return to main page"
            title="to main"
            :link="ROUTES.MAIN?.link"
          />
        </div>
      </div>
    </NuxtLayout>
  </div>
</template>

<script setup>
import { useRouter } from 'nuxt/app';
import { ROUTES } from './configs/data/routes';

defineProps({
  error: {
    type: Object,
    required: true,
  },
});

const {
  public: { isDev },
} = useRuntimeConfig();

onMounted(() => {
  if (!isDev) {
    const router = useRouter();
    router.replace({ path: '/' });
  }
});
</script>

<style scoped lang="scss">
.error {
  $parent: &;

  display: flex;
  flex-direction: column;
  min-height: 100svh;

  &:deep(#{$parent}__layout) {
    flex-grow: 1;
  }

  &__container {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 40px;
    align-items: center;
    justify-content: center;
  }

  &__button {
    min-width: 200px;
    max-width: fit-content;
    text-transform: capitalize;
  }
}
</style>
