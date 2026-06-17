export default defineNuxtPlugin({
  parallel: true,
  setup(nuxtApp) {
    nuxtApp.vueApp.directive('mask', {
      getSSRProps() {
        return {};
      },
    });
  },
});
