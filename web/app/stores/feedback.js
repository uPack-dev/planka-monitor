import { defineStore } from 'pinia';

export const useFeedbackStore = defineStore('feedback', {
  state: () => ({
    utmParams: {},
  }),
  actions: {
    setUtm(utms) {
      this.utmParams = utms;
    },

    sendLeadsData() {
      return Promise.resolve();
    },
  },
});
