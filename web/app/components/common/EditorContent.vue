<script setup>
defineProps({
  content: {
    type: String,
    default: '',
  },
});
</script>

<template>
  <div class="editor-content" v-html="content"></div>
</template>

<style lang="scss">
.editor-content {
  $text-font-size: 16;
  $text-font-size-adaptive: 15;

  color: $text-color-secondary;

  > * {
    &:first-child {
      margin-top: 0;
    }

    &:last-child {
      margin-bottom: 0;
    }
  }

  h1,
  h2,
  h3,
  h4,
  h5,
  h6 {
    color: $text-color-primary;
    text-transform: uppercase;
  }

  h1 {
    @include h2;

    margin-bottom: 44px;

    @include media-breakpoint-down(sm) {
      margin-bottom: 28px;
    }
  }

  h2 {
    @include h3;

    margin-bottom: 16px;

    @include media-breakpoint-down(sm) {
      margin-bottom: 16px;
    }
  }

  h3 {
    @include h4;

    margin-bottom: 16px;

    @include media-breakpoint-down(sm) {
      margin-bottom: 16px;
    }
  }

  h4 {
    @include h5;

    margin-bottom: 16px;

    @include media-breakpoint-down(sm) {
      margin-bottom: 16px;
    }
  }

  h5,
  h6 {
    @include h6;

    margin-bottom: 16px;

    @include media-breakpoint-down(sm) {
      margin-bottom: 16px;
    }
  }

  b,
  strong {
    color: $text-color-primary;
  }

  p,
  li,
  blockquote {
    @include s2('r', 's');
  }

  p {
    margin-bottom: 24px;

    @include media-breakpoint-down(sm) {
      margin-bottom: 24px;
    }
  }

  ol,
  ul {
    display: flex;
    flex-direction: column;
    gap: 24px;
    padding-left: 24px;
    margin: 24px 0;

    li {
      &::before {
        display: inline-block;
      }
    }
  }

  ol {
    counter-reset: li;

    li {
      counter-increment: li;

      &::before {
        width: 20px;
        margin-right: 9px;
        margin-left: -29px;
        font: inherit;
        color: $text-color-secondary;
        text-align: right;
        content: '.' counter(li);
        direction: rtl;

        @include media-breakpoint-down(sm) {
          width: 20px;
          margin-right: 9px;
          margin-left: -29px;
        }
      }
    }
  }

  ul {
    li {
      &::before {
        width: 6px;
        height: 6px;
        margin: 0 9px 2px -15px;
        content: '';
        background-color: $text-color-secondary;
        border-radius: 50%;

        @include media-breakpoint-down(sm) {
          width: 5px;
          height: 5px;
          margin: 0 9px 2px -15px;
        }
      }
    }
  }

  a,
  u {
    text-decoration: underline;
  }
}
</style>
