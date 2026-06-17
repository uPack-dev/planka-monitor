import { defineRule, configure } from 'vee-validate';

export default defineNuxtPlugin(() => {
  configure({
    validateOnInput: true,
  });

  defineRule('required', (value) => {
    if (!value || value.length === 0) {
      return 'This field is required';
    }
    return true;
  });

  defineRule('requiredCheckbox', (value) => {
    if (value !== true) {
      return 'You must agree to the terms';
    }
    return true;
  });

  // Updated to US Format
  defineRule('phone', (value) => {
    if (!value) {
      return 'Please enter a phone number';
    }

    // Static mask: +1 (XXX) XXX-XXXX
    const phoneRegex = /^\+1 \(\d{3}\) \d{3}-\d{4}$/;

    if (!phoneRegex.test(value)) {
      return 'Invalid phone number format';
    }

    return true;
  });

  defineRule('email', (value) => {
    if (!value) {
      return 'Please enter an email';
    }
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(value)) {
      return 'Please enter a valid email address';
    }
    return true;
  });

  defineRule('min', (value, [limit]) => {
    if (!value || value.length < limit) {
      return `Minimum length is ${limit} characters`;
    }
    return true;
  });

  defineRule('max', (value, [limit]) => {
    if (value && value.length > limit) {
      return `Maximum length is ${limit} characters`;
    }
    return true;
  });

  defineRule('alpha', (value) => {
    if (!value) return true;
    const alphaRegex = /^[a-zA-Zа-яА-Я]+$/;
    if (!alphaRegex.test(value)) {
      return 'This field should contain only letters';
    }
    return true;
  });

  defineRule('alpha_spaces', (value) => {
    if (!value) return true;
    const alphaSpacesRegex = /^[a-zA-Zа-яА-Я\s]+$/;
    if (!alphaSpacesRegex.test(value)) {
      return 'This field should contain only letters and spaces';
    }
    return true;
  });

  defineRule('alpha_num', (value) => {
    if (!value) return true;
    const alphaNumRegex = /^[a-zA-Zа-яА-Я0-9]+$/;
    if (!alphaNumRegex.test(value)) {
      return 'This field should contain only letters and numbers';
    }
    return true;
  });

  defineRule('length', (value, [exactLength]) => {
    if (!value) {
      return `This field must be exactly ${exactLength} characters`;
    }
    if (value.length !== parseInt(exactLength)) {
      return `This field must be exactly ${exactLength} characters`;
    }
    return true;
  });

  defineRule('digits', (value) => {
    if (!value) return true;
    const digitsRegex = /^\d+$/;
    if (!digitsRegex.test(value)) {
      return 'This field should contain only digits';
    }
    return true;
  });

  defineRule('url', (value) => {
    if (!value) return true;
    const urlRegex = /^(https?:\/\/)?([\w-]+(\.[\w-]+)+)(:\d{1,5})?(\/\S*)?$/i;
    if (!urlRegex.test(value)) {
      return 'Please enter a valid URL';
    }
    return true;
  });
});
