<template>
  <div class="layouts-default">
    <slot class-name="layouts-default__view" />

    <div id="modal-root" />

    <ClientOnly>
      <ModalsContainer />
    </ClientOnly>
  </div>
</template>

<script setup>
import { ModalsContainer } from 'vue-final-modal';
import { useScrollLock } from '@/composables/useScrollLock';

const scrollLock = useScrollLock();

defineProps({
  className: {
    type: String,
    default: undefined,
  },
});

onBeforeUnmount(() => {
  scrollLock.unlock();
});
</script>

<style scoped lang="scss">
.layouts-default {
  min-height: 100vh;
  min-height: 100dvh;
  background-color: $annonce-color-navy;

  &:deep(.layouts-default__view) {
    min-height: 100vh;
    min-height: 100dvh;
  }
}
</style>
