<template>
  <span
    class="mini-avatar"
    :class="[`mini-avatar--${size}`, { 'mini-avatar--bordered': bordered }]"
    :style="{ '--mini-avatar-bg': backgroundColor }"
    :title="displayName || undefined"
    aria-hidden="true"
  >
    <img
      v-if="imageUrl"
      :src="imageUrl"
      alt=""
      draggable="false"
      @error="avatarFailed = true"
    />
    <span v-else class="mini-avatar__letter">{{ fallbackLetter }}</span>
  </span>
</template>

<script setup>
const props = defineProps({
  user: {
    type: Object,
    default: null,
  },
  src: {
    type: String,
    default: '',
  },
  name: {
    type: String,
    default: '',
  },
  color: {
    type: String,
    default: '',
  },
  size: {
    type: String,
    default: 'md',
    validator: (value) => ['xs', 'sm', 'md', 'lg', 'xl', 'xxl'].includes(value),
  },
  bordered: {
    type: Boolean,
    default: false,
  },
});

const avatarFailed = shallowRef(false);

const avatarSource = computed(() =>
  String(
    props.src || props.user?.avatarUrl || props.user?.workspaceAvatarUrl || '',
  ).trim(),
);
const imageUrl = computed(() => (avatarFailed.value ? '' : avatarSource.value));
const displayName = computed(() =>
  String(
    props.name ||
      props.user?.name ||
      props.user?.displayName ||
      props.user?.workspaceDisplayName ||
      props.user?.username ||
      props.user?.workspaceUsername ||
      props.user?.telegramUsername ||
      props.user?.initials ||
      '',
  ).trim(),
);
const fallbackLetter = computed(() => firstLetter(displayName.value));
const backgroundColor = computed(
  () => props.color || props.user?.color || colorFromName(displayName.value),
);

watch(avatarSource, () => {
  avatarFailed.value = false;
});

function firstLetter(value) {
  const text = String(value || '')
    .trim()
    .replace(/^@+/, '');
  return Array.from(text)[0]?.toLocaleUpperCase('uk-UA') || '?';
}

function colorFromName(value) {
  const palette = ['#275f88', '#d94d67', '#2e8c5a', '#5995c1', '#0b304b'];
  const seed = Array.from(String(value || '?')).reduce(
    (sum, char) => sum + char.charCodeAt(0),
    0,
  );
  return palette[seed % palette.length];
}
</script>

<style scoped lang="scss">
.mini-avatar {
  --mini-avatar-size: 28px;
  --mini-avatar-font-size: 12px;

  display: grid;
  flex: 0 0 var(--mini-avatar-size);
  place-items: center;
  width: var(--mini-avatar-size);
  height: var(--mini-avatar-size);
  overflow: hidden;
  font-size: var(--mini-avatar-font-size);
  font-weight: 800;
  line-height: 1;
  color: #ffffff;
  user-select: none;
  background-color: var(--mini-avatar-bg);
  border-radius: 50%;
  box-shadow: inset 0 0 0 1px rgba(#ffffff, 0.16);
  transition:
    box-shadow $time-fast $ease-out,
    transform $time-fast $ease-out;

  &__letter {
    display: grid;
    place-items: center;
    width: 100%;
    height: 100%;
    animation: mini-soft-pop 180ms $ease-out both;
  }

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: inherit;
    animation: mini-rise-in 200ms $ease-out both;
  }

  &--xs {
    --mini-avatar-size: 18px;
    --mini-avatar-font-size: 9px;
  }

  &--sm {
    --mini-avatar-size: 26px;
    --mini-avatar-font-size: 12px;
  }

  &--lg {
    --mini-avatar-size: 42px;
    --mini-avatar-font-size: 16px;
  }

  &--xl {
    --mini-avatar-size: 48px;
    --mini-avatar-font-size: 18px;
  }

  &--xxl {
    --mini-avatar-size: 74px;
    --mini-avatar-font-size: 28px;
  }

  &--bordered {
    border: 1px solid #fffdfa;
  }
}

@media (prefers-reduced-motion: reduce) {
  .mini-avatar,
  .mini-avatar__letter,
  .mini-avatar img {
    transition-duration: 1ms;
    animation: none;
  }
}
</style>
