<template>
  <div class="welcome">
    <header class="welcome__top">
      <MiniProgressDots :active="1" />
      <button class="welcome__skip" type="button" @click="$emit('skip')">
        Пропустити
      </button>
    </header>

    <div class="welcome__logo">A</div>

    <div class="welcome__copy">
      <h1>Annonce —<br />ваш Workspace<br />під рукою.</h1>
      <p>
        Дедлайни, призначення, коментарі з Planka — на шкалі часу. Тільки те, що
        стосується вас.
      </p>
    </div>

    <div class="welcome__features">
      <MiniFeatureRow
        :style="{ '--motion-delay': '0ms' }"
        icon="clock-mini"
        title="Шкала часу"
        text="Задачі впорядковані від «зараз» до «пізніше»"
      />
      <MiniFeatureRow
        :style="{ '--motion-delay': '35ms' }"
        icon="lock-mini"
        title="Лише ваше"
        text="Сповіщення про згадки, призначення та коментарі"
      />
      <MiniFeatureRow
        :style="{ '--motion-delay': '70ms' }"
        icon="workspace-mini"
        title="Дієте з міні-апки"
        text="Виконано, коментар, відкрити Workspace"
      />
    </div>

    <UiButton
      class="welcome__button"
      title="Почати  →"
      :theme="BUTTON_THEME.ANNONCE"
      :size="BUTTON_SIZE.LG"
      @click="$emit('start')"
    />
  </div>
</template>

<script setup>
import { BUTTON_SIZE, BUTTON_THEME } from '@/configs/uiButtonOptions';

defineEmits(['skip', 'start']);
</script>

<style scoped lang="scss">
.welcome {
  display: flex;
  flex-direction: column;
  min-height: 100%;

  &__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  &__skip {
    font-size: 13.5px;
    line-height: 18px;
    color: $annonce-color-muted;
    opacity: 0.95;
    transition:
      color $time-fast $ease-out,
      opacity $time-fast $ease-out,
      transform $time-fast $ease-out;
  }

  &__logo {
    display: grid;
    place-items: center;
    width: 64px;
    height: 64px;
    margin-top: 36px;
    font-family: $annonce-font-brand;
    font-size: 26px;
    line-height: 30px;
    color: $annonce-color-blue;
    background-color: $annonce-color-yellow;
    border-radius: 20px;
    box-shadow: 0 18px 42px rgba(#031d31, 0.22);
    animation: mini-soft-pop 240ms $ease-out both;
  }

  &__copy {
    margin-top: 50px;
    animation: mini-rise-in 240ms $ease-out both;
    animation-delay: 18ms;

    h1 {
      font-size: 24px;
      font-weight: 700;
      line-height: 1.25;
      color: $annonce-color-text;
    }

    p {
      max-width: 354px;
      margin-top: 20px;
      font-size: 15px;
      line-height: 22px;
      color: rgba($annonce-color-muted-strong, 0.86);
    }
  }

  &__features {
    display: grid;
    gap: 20px;
    margin-top: 50px;
  }

  &__button {
    margin-top: 100px;
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 18ms;
  }
}

@media (hover: hover) {
  .welcome__skip:hover {
    color: $annonce-color-text;
    opacity: 1;
    transform: translate3d(0, -1px, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .welcome__logo,
  .welcome__copy,
  .welcome__button {
    animation: none;
  }

  .welcome__skip {
    transition-duration: 1ms;
  }
}
</style>
