<script setup>
import { useModal } from 'vue-final-modal';

const emit = defineEmits(['close']);

const modalStatus = useModal({
  component: resolveComponent('LazyModalsStatus'),
});

const {
  public: { isDev },
} = useRuntimeConfig();

const token = ref('');

const nameField = ref('');
const phoneField = ref('');
const emailField = ref('');
const commentField = ref('');
const formRef = ref(null);
const isLoading = ref(false);

const onSubmit = async () => {
  const isValid = await formRef.value?.validate();

  if (!isValid) return;
  if (!isDev && !token.value) return;

  function cleanPhone(phone) {
    const digits = phone.replace(/\D/g, '');
    return `+${digits}`;
  }

  try {
    isLoading.value = true;

    const formData = {
      token: token.value,
      name: nameField.value,
      phone: cleanPhone(phoneField.value),
      email: emailField.value,
      description: commentField.value || '',
    };

    if (!isDev) await useRequest('/order', { method: 'post', body: formData });
    else alert(JSON.stringify(formData));

    modalStatus.patchOptions({
      attrs: {
        title: 'Thank you!',
        status: 'success',
        description: 'Our manager will contact you shortly.',
      },
    });
    setTimeout(async () => {
      await modalStatus.open();
    }, 500);

    nameField.value = '';
    phoneField.value = '';
    emailField.value = '';
    commentField.value = '';
    emit('close');
  } catch {
    emit('close');
    modalStatus.patchOptions({
      attrs: {
        title: 'Error',
        status: 'error',
        description: 'Something went wrong, please try again later.',
      },
    });
    setTimeout(async () => {
      await modalStatus.open();
    }, 500);
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <VeeForm
    ref="formRef"
    v-slot="{ failed }"
    class="ui-form"
    @submit="onSubmit()"
  >
    <div class="ui-form__container">
      <div class="ui-form__box">
        <UiInput
          v-model="nameField"
          name="Name"
          rules="required|alpha_spaces"
          label="Name"
        />

        <UiInput
          v-model="phoneField"
          type="tel"
          name="Phone"
          :mask="{
            mask: '+1 (###) ###-####',
            eager: true,
          }"
          rules="phone|required"
          label="Phone"
        />

        <UiInput
          v-model="emailField"
          name="Email"
          rules="email|required"
          label="Email"
        />

        <UiTextArea
          v-model="commentField"
          name="Info"
          label="Brief description of the problem"
        />
      </div>
    </div>

    <ACrossFade>
      <UiButton
        v-if="token"
        title="Order a service"
        type="submit"
        :disabled="failed || isLoading"
        class="ui-form__button"
        icon="arrow-right"
      />

      <NuxtTurnstile v-else v-model="token" />
    </ACrossFade>
  </VeeForm>
</template>

<style lang="scss" scoped>
.ui-form {
  display: flex;
  flex-direction: column;
  gap: 30px;
  align-items: center;
  text-align: left;

  @include media-breakpoint-down(sm) {
    gap: 20px;
  }

  &__container {
    width: 350px;

    @include media-breakpoint-down(sm) {
      width: 320px;
      max-width: max(calc(100vw - 40px), 295px);
    }
  }

  &__box {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  &__button {
    width: max-content;
    min-width: 320px;
    margin-inline: auto;
  }
}
</style>
