<template>
  <MiniAppShell theme="light">
    <div class="empty-workspace">
      <div class="empty-workspace__header">
        <div class="empty-workspace__copy">
          <span>ANNONCE · ADMIN</span>
          <strong>Шкала задач</strong>
        </div>
        <div class="empty-workspace__actions">
          <button type="button" aria-label="Пошук" disabled>
            <CIcon name="search-mini" />
          </button>
          <button type="button" aria-label="Меню" disabled>
            <CIcon name="more-mini" />
          </button>
        </div>
        <div class="empty-workspace__tabs" aria-hidden="true">
          <span>Список</span>
          <span>Календар</span>
        </div>
      </div>
      <div class="empty-workspace__body">
        <span class="empty-workspace__icon">
          <CIcon name="link" />
        </span>
        <h1>Поки що порожньо</h1>
        <p>
          Привʼяжіть свого Workspace-юзера, щоб тут зʼявились дедлайни ваших
          карток та задач.
        </p>
        <UiButton
          title="Привʼязати акаунт  →"
          :theme="BUTTON_THEME.ANNONCE"
          @click="$emit('bind')"
        />
      </div>
      <MiniBottomNav active="tasks" />
    </div>
  </MiniAppShell>
</template>

<script setup>
import { BUTTON_THEME } from '@/configs/uiButtonOptions';

defineEmits(['bind']);
</script>

<style scoped lang="scss">
.empty-workspace {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  margin: -24px;

  &__icon {
    display: grid;
    place-items: center;
    width: 90px;
    height: 90px;
    color: $annonce-color-blue;
    background-color: #e1edf5;
    border: 1.5px solid #b0c9da;
    border-radius: 26px;
    box-shadow: 0 16px 36px rgba(#0b304b, 0.1);
    animation: mini-soft-pop 230ms $ease-out both;

    :deep(svg) {
      width: 44px;
      height: 44px;
    }
  }

  &__header {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 12px;
    align-items: start;
    min-height: 126px;
    padding: calc(var(--tg-safe-top, 0px) + 20px) 18px 10px;
    color: $annonce-color-text;
    background-color: $annonce-color-navy;
    animation: mini-rise-in 220ms $ease-out both;
  }

  &__copy {
    display: grid;
    gap: 7px;

    span {
      font-size: 12px;
      font-weight: 700;
      color: rgba($annonce-color-muted-strong, 0.66);
    }

    strong {
      font-size: 20px;
      line-height: 1.2;
    }
  }

  &__actions {
    display: flex;
    gap: 5px;

    button {
      display: grid;
      place-items: center;
      width: 44px;
      height: 44px;
      color: $annonce-color-text;
      background-color: rgba($color-white, 0.16);
      border: 1.5px solid rgba($color-white, 0.26);
      border-radius: 14px;
      opacity: 0.28;
    }

    :deep(svg) {
      width: 20px;
      height: 20px;
    }
  }

  &__tabs {
    display: flex;
    grid-column: 1 / -1;
    width: 143px;
    padding: 3px;
    background-color: #265577;
    border: 1.5px solid rgba($color-white, 0.26);
    border-radius: 17px;

    span {
      display: grid;
      place-items: center;
      min-width: 62px;
      height: 28px;
      font-size: 12px;
      font-weight: 700;
      line-height: 1;
      color: #b4c8d2;
      border-radius: 14px;

      &:first-child {
        color: $annonce-color-navy;
        background-color: $annonce-color-yellow;
      }
    }
  }

  &__body {
    display: grid;
    gap: 18px;
    justify-items: center;
    padding: 24px 24px calc(var(--tg-safe-bottom, 0px) + 128px);
    margin-block: auto;
    text-align: center;
    animation: mini-rise-in 240ms $ease-out both;
    animation-delay: 45ms;

    h1 {
      font-size: 30px;
      line-height: 1.1;
    }

    p {
      max-width: 286px;
      font-size: 15px;
      line-height: 1.45;
      color: $annonce-color-ink-muted;
    }
  }
}

@media (prefers-reduced-motion: reduce) {
  .empty-workspace__icon,
  .empty-workspace__header,
  .empty-workspace__body {
    animation: none;
  }
}
</style>
