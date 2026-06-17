<script setup></script>

<template>
  <div class="ui-website-developer">
    <div class="ui-website-developer__wrapper">
      <div class="ui-website-developer__title hidden-mobile">
        <p class="s2-m-s">Website developer</p>
      </div>

      <CIcon name="dev" class="ui-website-developer__icon" />
    </div>

    <UiPrivacyPolicy class="hidden-mobile" />
  </div>
</template>

<style lang="scss" scoped>
.ui-website-developer {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-end;
  align-self: stretch;

  @include media-breakpoint-down(sm) {
    align-items: center;
    align-self: auto;
  }

  &__icon {
    width: 70px;
    height: 70px;
    border: 2px solid $color-white;
  }

  &__wrapper {
    display: inherit;
    flex-direction: inherit;
    gap: 12px;
    align-items: inherit;
  }

  &__title {
    p::after {
      content: '↓';
    }
  }
}
</style>
