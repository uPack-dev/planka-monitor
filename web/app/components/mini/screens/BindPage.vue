<template>
  <VeeForm class="bind" @submit="$emit('submit')">
    <MiniProgressDots :active="2" class="bind__progress" />

    <p class="bind__step">КРОК 2 · ПРИВʼЯЗКА</p>
    <h1 class="bind__title">Який ви у Workspace?</h1>
    <p class="bind__text">
      Щоб писати персонально, треба знати ваш planka-username. Telegram і
      Workspace можуть відрізнятись.
    </p>

    <button
      v-if="suggestedWorkspace"
      type="button"
      class="bind__choice bind__match"
      :class="{
        'bind__choice--active': selectedWorkspaceMode === 'suggested',
      }"
      :aria-pressed="selectedWorkspaceMode === 'suggested'"
      @click="emit('select-suggested-workspace')"
    >
      <MiniAvatar
        class="bind__avatar"
        :src="avatarSource"
        :name="displayName"
        color="#ffc261"
        size="xl"
      />
      <div class="bind__match-copy">
        <span>знайдено за username:</span>
        <strong>{{ displayName }}</strong>
        <small
          >@{{ telegramUsername }} → planka: {{ suggestedWorkspace }}</small
        >
      </div>
      <span class="bind__chip">Знайдено</span>
    </button>

    <button
      v-if="suggestedWorkspace"
      type="button"
      class="bind__choice bind__alternate"
      :class="{
        'bind__choice--active': selectedWorkspaceMode === 'manual',
      }"
      :aria-pressed="selectedWorkspaceMode === 'manual'"
      @click="emit('select-manual-workspace')"
    >
      <span>Це не я — використати інший username</span>
    </button>

    <label class="bind__label" for="workspace-username"
      >WORKSPACE USERNAME</label
    >
    <UiInput
      id="workspace-username"
      class="bind__input"
      :class="{
        'bind__input--active': selectedWorkspaceMode === 'manual',
      }"
      :model-value="workspaceInput"
      name="workspaceUsername"
      :theme="INPUT_THEME.ANNONCE"
      placeholder="username"
      @click="emit('activate-manual-workspace')"
      @update:model-value="emit('update:workspace-input', $event)"
    >
      <span class="bind__input-prefix">planka:</span>
    </UiInput>

    <p v-if="bindingError" class="screen-error">{{ bindingError }}</p>

    <UiButton
      class="bind__submit"
      title="Далі  →"
      type-button="submit"
      :theme="BUTTON_THEME.ANNONCE"
      :size="BUTTON_SIZE.LG"
      :is-loading="isLoading"
    />
  </VeeForm>
</template>

<script setup>
import { BUTTON_SIZE, BUTTON_THEME } from '@/configs/uiButtonOptions';
import { INPUT_THEME } from '@/configs/uiInputOptions';

defineProps({
  suggestedWorkspace: {
    type: String,
    default: '',
  },
  avatarSource: {
    type: String,
    default: '',
  },
  displayName: {
    type: String,
    required: true,
  },
  telegramUsername: {
    type: String,
    default: '',
  },
  workspaceInput: {
    type: String,
    required: true,
  },
  selectedWorkspaceMode: {
    type: String,
    required: true,
    validator(value) {
      return ['suggested', 'manual'].includes(value);
    },
  },
  bindingError: {
    type: String,
    default: '',
  },
  isLoading: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits([
  'activate-manual-workspace',
  'select-manual-workspace',
  'select-suggested-workspace',
  'submit',
  'update:workspace-input',
]);
</script>

<style scoped lang="scss">
.bind {
  display: flex;
  flex-direction: column;
  min-height: 100%;

  &__progress {
    margin-bottom: 91px;
  }

  &__step {
    font-size: 13px;
    font-weight: 700;
    line-height: 1.4;
    color: rgba($annonce-color-muted, 0.9);
    animation: mini-rise-in 220ms $ease-out both;
  }

  &__title {
    margin-top: 10px;
    font-size: 24px;
    font-weight: 700;
    line-height: 30px;
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 18ms;
  }

  &__text {
    max-width: 330px;
    margin-top: 20px;
    font-size: 15.5px;
    line-height: 23px;
    color: rgba($annonce-color-muted, 0.96);
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 35ms;
  }

  &__choice {
    width: 100%;
    font-family: inherit;
    color: inherit;
    text-align: left;
    appearance: none;
    cursor: pointer;
    border: 1.5px solid rgba($annonce-color-muted, 0.55);
    transition:
      background-color $time-normal ease,
      border-color $time-normal ease,
      box-shadow $time-normal ease,
      transform $time-normal ease;

    &:active {
      transform: scale(0.985);
    }

    &--active {
      border-color: $annonce-color-yellow;
      box-shadow: 0 0 0 1px rgba($annonce-color-yellow, 0.45);
    }
  }

  &__match {
    display: grid;
    grid-template-columns: 48px 1fr auto;
    gap: 14px;
    align-items: center;
    min-height: 118px;
    padding: 16px;
    margin-top: 31px;
    background-color: rgba($color-white, 0.1);
    border-radius: 16px;
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 20ms;

    @include hover {
      background-color: rgba($color-white, 0.13);
      border-color: rgba($annonce-color-yellow, 0.85);
    }
  }

  &__match-copy {
    display: grid;
    gap: 2px;
    min-width: 0;

    span {
      font-size: 12px;
      color: rgba($annonce-color-muted-strong, 0.7);
    }

    strong {
      overflow: hidden;
      text-overflow: ellipsis;
      font-family: $annonce-font-brand;
      font-size: 17px;
      font-weight: 400;
      line-height: 21px;
      white-space: nowrap;
    }

    small {
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 12px;
      color: rgba($annonce-color-text, 0.86);
      white-space: nowrap;
    }
  }

  &__chip {
    padding: 5px 9px;
    font-size: 13px;
    font-weight: 700;
    line-height: 16px;
    color: $annonce-color-blue;
    background-color: #e4f5f7;
    border: 1.5px solid #b8dde5;
    border-radius: 16px;
  }

  &__alternate {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 48px;
    padding: 13px 15px;
    margin-top: 20px;
    font-size: 14px;
    font-weight: 700;
    line-height: 20px;
    color: $annonce-color-muted;
    text-align: center;
    background-color: rgba($color-white, 0.08);
    border-radius: 14px;
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 25ms;

    @include hover {
      color: $annonce-color-text;
      background-color: rgba($color-white, 0.13);
      border-color: rgba($annonce-color-yellow, 0.85);
    }
  }

  &__label {
    margin-top: 50px;
    margin-bottom: 10px;
    font-size: 13.5px;
    font-weight: 600;
    line-height: 18px;
    color: rgba($annonce-color-muted, 0.9);
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 15ms;
  }

  &__input {
    animation: mini-rise-in 220ms $ease-out both;
    animation-delay: 35ms;

    :deep(.ui-input) {
      min-height: 48px;
      padding: 15px;
      transition:
        border-color $time-normal $ease-out,
        box-shadow $time-normal $ease-out,
        transform $time-normal $ease-out;
    }

    :deep(.ui-input__field) {
      padding-left: 58px;
      font-size: 14.5px;
      font-weight: 400;
      line-height: 18px;
    }

    &--active {
      :deep(.ui-input) {
        border-color: $annonce-color-yellow;
        box-shadow: 0 0 0 1px rgba($annonce-color-yellow, 0.45);
      }
    }
  }

  &__input-prefix {
    position: absolute;
    top: 50%;
    left: 15px;
    font-size: 14.5px;
    font-weight: 500;
    line-height: 18px;
    color: $annonce-color-muted;
    pointer-events: none;
    translate: 0 -50%;
  }

  &__submit {
    margin-top: 90px;
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
  .bind__step,
  .bind__title,
  .bind__text,
  .bind__match,
  .bind__alternate,
  .bind__label,
  .bind__input,
  .bind__submit {
    animation: none;
  }

  .bind__choice,
  .bind__input :deep(.ui-input) {
    transition-duration: 1ms;
  }
}
</style>
