import { MasonryWall } from '@yeger/vue-masonry-wall';

export default defineNuxtPlugin({
  parallel: true,
  setup(nuxtApp) {
    nuxtApp.vueApp.component('MasonryWall', MasonryWall);
  },
});
