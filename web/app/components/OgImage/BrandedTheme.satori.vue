<script setup>
import { OG_IMAGE_COLORS } from '@/constants/colors';

defineProps({
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    default: '',
  },
  backgroundImage: {
    type: String,
    default: '',
  },
  icon: {
    type: String,
    default: '',
  },
});

const cardStyle = {
  position: 'relative',
  display: 'flex',
  flexDirection: 'column',
  justifyContent: 'flex-end',
  width: '100%',
  height: '100%',
  padding: '48px',
  color: OG_IMAGE_COLORS.textInvert,
  backgroundColor: OG_IMAGE_COLORS.background,
};

const imageStyle = {
  position: 'absolute',
  top: 0,
  left: 0,
  width: '100%',
  height: '100%',
  objectFit: 'cover',
};

const iconStyle = {
  position: 'absolute',
  top: '48px',
  right: '48px',
  width: '240px',
  height: '150px',
  fontSize: '56px',
  fontWeight: 700,
  lineHeight: 1,
  textAlign: 'right',
};

const overlayStyle = {
  position: 'absolute',
  top: 0,
  left: 0,
  width: '100%',
  height: '100%',
  backgroundColor: OG_IMAGE_COLORS.overlay,
};

const contentStyle = {
  position: 'relative',
  display: 'flex',
  flexDirection: 'column',
  gap: '16px',
  height: 'auto',
};

const titleStyle = {
  margin: 0,
  fontFamily: 'Arial, sans-serif',
  fontSize: '60px',
  fontWeight: 900,
  lineHeight: 1.1,
};

const descriptionStyle = {
  margin: 0,
  fontFamily: 'Arial, sans-serif',
  fontSize: '30px',
  lineHeight: 1.4,
  color: OG_IMAGE_COLORS.textMutedInvert,
};
</script>

<template>
  <div :style="cardStyle">
    <img
      v-if="backgroundImage"
      :src="backgroundImage"
      alt=""
      :style="imageStyle"
    />
    <div :style="overlayStyle" />
    <div v-if="icon" :style="iconStyle">
      {{ icon }}
    </div>
    <div :style="contentStyle">
      <h1 :style="titleStyle">
        {{ title }}
      </h1>
      <p v-if="description" :style="descriptionStyle">
        {{ description }}
      </p>
    </div>
  </div>
</template>
