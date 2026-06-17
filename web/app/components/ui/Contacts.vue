<script setup>
import { useGlobalStore } from '@/stores/global';
import { escapeNonNumeric } from '@/utils/helpers';

const globalStore = useGlobalStore();
const { phone, email } = globalStore.contacts;

const items = [];

if (phone?.length)
  items.push(
    ...phone.map((phone) => ({
      title: phone,
      link: `tel:+${escapeNonNumeric(phone)}`,
      icon: 'phone',
    })),
  );

if (email?.length)
  items.push(
    ...email.map((email) => ({
      title: email,
      link: `mailto:${email}`,
      icon: 'email',
    })),
  );
</script>

<template>
  <ul class="ui-contacts">
    <li
      v-for="({ title, link, icon }, index) in items"
      :key="index"
      class="ui-contacts__item"
    >
      <CLinkTag class="ui-contacts__link" :link="link">
        <CIcon class="ui-contacts__icon" :name="icon" />

        <span class="ui-contacts__font i1-m-s">{{ title }}</span>
      </CLinkTag>
    </li>
  </ul>
</template>

<style lang="scss" scoped>
.ui-contacts {
  &__font {
    line-height: 160%;
  }

  &__icon {
    width: 30px;
    height: 30px;
  }

  &__link {
    display: inline-flex;
    gap: 10px;
    align-items: center;
    transition: color $time-normal $ease-out;

    @include hover-active-focus {
      color: $actions-color-red;
    }
  }
}
</style>
