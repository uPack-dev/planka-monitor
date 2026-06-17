<template>
  <component :is="component" v-bind="props">
    <slot />
  </component>
</template>

<script setup>
import { NuxtLink } from '#components';

const initialProps = defineProps({
  link: {
    type: [String, Boolean, Object],
    default: undefined,
  },
  target: {
    type: String,
    default: undefined,
  },
  type: {
    type: String,
    default: undefined,
  },
  ariaLabel: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
});

const isExternal = (link) => {
  return typeof link === 'string' && /^https?:\/\//.test(link);
};

const isInternal = (link) => {
  return (
    ((typeof link === 'object' && link !== null) ||
      (typeof link === 'string' && link.length > 0)) &&
    !isExternal(link)
  );
};

const component = computed(() => {
  if (isExternal(initialProps.link)) {
    return 'a';
  } else if (isInternal(initialProps.link)) {
    return NuxtLink;
  } else {
    return 'button';
  }
});

const props = computed(() => {
  const props = {};

  if (initialProps.ariaLabel) props['aria-label'] = initialProps.ariaLabel;

  if (initialProps.disabled) {
    props.tabindex = -1;
    props['aria-disabled'] = true;
  } else {
    props.tabindex = 0;
  }

  switch (component.value) {
    case 'a':
      props.href = initialProps.link;
      props.target = initialProps.target ?? '_blank';
      props.rel = 'nofollow noopener noreferrer';
      props.draggable = false;
      break;
    case NuxtLink:
      props.to = initialProps.link;
      props.draggable = false;
      break;
    case 'button':
      props.type = initialProps.type ?? 'button';
      break;
  }

  return props;
});
</script>
