<script setup>
import { normalizeMediaPath } from '@/utils/helpers';

const props = defineProps({
  src: {
    type: String,
    required: true,
  },
  srcset: {
    type: String,
    default: undefined,
  },
  alt: {
    type: String,
    default: '',
  },
  loading: {
    type: String,
    default: 'lazy',
    validator: (value) => ['lazy', 'eager'].includes(value),
  },
  fetchPriority: {
    type: String,
    default: undefined,
    validator: (value) => ['high', 'low', 'auto'].includes(value),
  },
});

const fetchPriorityComputed = computed(() => {
  if (props.fetchPriority) return props.fetchPriority;
  switch (props.loading) {
    case 'lazy': {
      return 'low';
    }
    case 'eager': {
      return 'high';
    }
    default: {
      return undefined;
    }
  }
});

const imageSrc = computed(() => {
  return normalizeMediaPath(props.src);
});
</script>

<template>
  <img
    :src="imageSrc"
    :srcset="srcset"
    :aria-hidden="!alt || undefined"
    :alt="alt"
    :loading="loading"
    :fetchpriority="fetchPriorityComputed"
  />
</template>
