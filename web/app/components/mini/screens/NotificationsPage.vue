<template>
  <div class="notifications">
    <MiniProgressDots :active="3" />
    <p class="notifications__step">КРОК 3 · СПОВІЩЕННЯ</p>
    <h1>Про що писати у Telegram?</h1>
    <p class="notifications__text">
      Тільки персональні події, де є ваше імʼя. Загальний канал залишається у
      груповому чаті.
    </p>

    <div class="notifications__list">
      <MiniToggleCard
        v-for="(item, itemIndex) in preferenceCards"
        :key="item.key"
        :style="{ '--motion-delay': `${itemIndex * 24}ms` }"
        :model-value="preferences[item.key]"
        :icon="item.icon"
        :title="item.title"
        :text="item.text"
        @update:model-value="$emit('update-preference', item.key, $event)"
      />
    </div>

    <p v-if="errorCode" class="screen-error">
      {{ preferenceError }}
    </p>

    <UiButton
      class="notifications__button"
      title="Готово"
      :theme="BUTTON_THEME.ANNONCE"
      :size="BUTTON_SIZE.LG"
      :is-loading="isLoading"
      @click="$emit('finish')"
    />

    <p class="notifications__hint">Можна змінити будь-коли в Профілі.</p>
  </div>
</template>

<script setup>
import { BUTTON_SIZE, BUTTON_THEME } from '@/configs/uiButtonOptions';

defineProps({
  preferenceCards: {
    type: Array,
    required: true,
  },
  preferences: {
    type: Object,
    required: true,
  },
  errorCode: {
    type: String,
    default: '',
  },
  preferenceError: {
    type: String,
    required: true,
  },
  isLoading: {
    type: Boolean,
    required: true,
  },
});

defineEmits(['update-preference', 'finish']);
</script>

<style scoped lang="scss">
.notifications {
  display: flex;
  flex-direction: column;
  min-height: 100%;

  &__step {
    margin-top: 88px;
    font-size: 13px;
    font-weight: 700;
    line-height: 18px;
    color: rgba($annonce-color-muted, 0.9);
    animation: mini-rise-in 220ms $ease-out both;
  }

  &__text {
    margin-top: 20px;
    font-size: 15.5px;
    line-height: 23px;
    color: rgba($annonce-color-muted, 0.96);
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 35ms;
  }

  &__list {
    display: grid;
    gap: 10px;
    margin-top: 31px;
  }

  &__button {
    margin-top: 90px;
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 35ms;
  }

  &__hint {
    margin-top: 12px;
    font-size: 13px;
    line-height: 18px;
    color: rgba($annonce-color-muted, 0.78);
    text-align: center;
  }

  h1 {
    margin-top: 10px;
    font-size: 24px;
    font-weight: 700;
    line-height: 30px;
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 18ms;
  }
}

.screen-error {
  font-size: 13px;
  line-height: 1.35;
  color: $annonce-color-error;
  text-align: center;
}

@media (prefers-reduced-motion: reduce) {
  .notifications__step,
  .notifications__text,
  .notifications__button,
  .notifications h1 {
    animation: none;
  }
}
</style>
