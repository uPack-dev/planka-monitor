<template>
  <section
    class="mini-shell"
    :class="{
      'mini-shell--dark': theme === 'dark',
      'mini-shell--light': theme === 'light',
      [`mini-shell--variant-${variant}`]: variant !== 'default',
      [`mini-shell--glow-${glowVariant}`]: glowVariant !== 'default',
    }"
  >
    <div
      v-if="theme === 'dark' && glowVariant !== 'none'"
      class="mini-shell__glow mini-shell__glow--primary"
    />
    <div
      v-if="theme === 'dark' && glowVariant !== 'none'"
      class="mini-shell__glow mini-shell__glow--secondary"
    />
    <div
      v-if="theme === 'dark' && glowVariant !== 'none'"
      class="mini-shell__glow mini-shell__glow--tertiary"
    />

    <div class="mini-shell__content">
      <slot />
    </div>
  </section>
</template>

<script setup>
defineProps({
  theme: {
    type: String,
    default: 'dark',
    validator(value) {
      return ['dark', 'light'].includes(value);
    },
  },
  variant: {
    type: String,
    default: 'default',
    validator(value) {
      return ['default', 'onboarding'].includes(value);
    },
  },
  glowVariant: {
    type: String,
    default: 'default',
    validator(value) {
      return ['default', 'welcome', 'bind', 'notifications', 'none'].includes(
        value,
      );
    },
  },
});
</script>

<style scoped lang="scss">
.mini-shell {
  position: relative;
  width: 100%;
  min-height: 100vh;
  min-height: 100dvh;
  min-height: var(--tg-viewport-height, 100dvh);
  overflow: hidden;
  font-family: Inter, sans-serif;
  isolation: isolate;

  &__content {
    position: relative;
    z-index: 1;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    min-height: 100vh;
    min-height: 100dvh;
    min-height: var(--tg-viewport-height, 100dvh);
    padding: calc(var(--tg-safe-top, 0px) + 24px) 24px
      calc(var(--tg-safe-bottom, 0px) + 24px);
  }

  &__glow {
    position: absolute;
    z-index: 0;
    width: 230px;
    height: 230px;
    pointer-events: none;
    background-color: rgba($annonce-color-yellow, 0.22);
    border-radius: 50%;

    &--primary {
      top: -60px;
      left: -70px;
    }

    &--secondary {
      right: -72px;
      bottom: -64px;
      background-color: rgba(#d85176, 0.2);
    }

    &--tertiary {
      display: none;
    }
  }

  &--dark {
    color: $annonce-color-text;
    background-color: $annonce-color-navy;
  }

  &--light {
    color: $annonce-color-ink;
    background-color: $annonce-color-cream;
  }

  &--variant-onboarding {
    background:
      radial-gradient(
        circle at 52% -10%,
        rgba($annonce-color-blue-soft, 0.24),
        transparent 35%
      ),
      $annonce-color-navy;

    .mini-shell {
      &__content {
        padding: calc(var(--tg-safe-top, 0px) + 24px) 22.5px
          calc(var(--tg-safe-bottom, 0px) + 24px);
      }

      &__glow {
        display: block;
        border: 1px solid rgba($color-white, 0.04);
        mix-blend-mode: screen;
        filter: saturate(1.08);
        transform: translate3d(0, 0, 0);
        will-change: transform, opacity;

        &--primary {
          top: -92px;
          left: -74px;
          width: 270px;
          height: 270px;
          background: radial-gradient(
            circle at 34% 35%,
            rgba($annonce-color-yellow, 0.28),
            rgba(#86a6bd, 0.33) 54%,
            rgba(#86a6bd, 0.18) 100%
          );
          animation: onboarding-orb-primary 28s $ease-in-out-sine infinite;
        }

        &--secondary {
          top: 88px;
          right: -88px;
          bottom: auto;
          width: 222px;
          height: 222px;
          background: radial-gradient(
            circle at 62% 38%,
            rgba(#d85176, 0.18),
            rgba(#8d80b4, 0.38) 62%,
            rgba(#8d80b4, 0.2) 100%
          );
          animation: onboarding-orb-secondary 32s $ease-in-out-sine infinite;
        }

        &--tertiary {
          top: auto;
          bottom: -104px;
          left: -78px;
          width: 250px;
          height: 250px;
          background: radial-gradient(
            circle at 42% 40%,
            rgba($annonce-color-yellow, 0.14),
            rgba(#78adc2, 0.38) 58%,
            rgba(#78adc2, 0.2) 100%
          );
          animation: onboarding-orb-tertiary 34s $ease-in-out-sine infinite;
        }
      }
    }
  }
}

@media (prefers-reduced-motion: reduce) {
  .mini-shell--variant-onboarding .mini-shell__glow {
    animation: none;
  }
}

@keyframes onboarding-orb-primary {
  0%,
  100% {
    opacity: 0.9;
    transform: translate3d(0, 0, 0) scale(1);
  }

  34% {
    opacity: 0.82;
    transform: translate3d(8px, 6px, 0) scale(0.99);
  }

  68% {
    opacity: 0.94;
    transform: translate3d(-4px, 12px, 0) scale(1.02);
  }
}

@keyframes onboarding-orb-secondary {
  0%,
  100% {
    opacity: 0.78;
    transform: translate3d(0, 0, 0) scale(1);
  }

  42% {
    opacity: 0.9;
    transform: translate3d(-10px, 7px, 0) scale(1.02);
  }

  72% {
    opacity: 0.76;
    transform: translate3d(4px, -6px, 0) scale(0.99);
  }
}

@keyframes onboarding-orb-tertiary {
  0%,
  100% {
    opacity: 0.82;
    transform: translate3d(0, 0, 0) scale(1);
  }

  38% {
    opacity: 0.74;
    transform: translate3d(12px, -8px, 0) scale(1.02);
  }

  76% {
    opacity: 0.9;
    transform: translate3d(4px, 4px, 0) scale(0.99);
  }
}
</style>
