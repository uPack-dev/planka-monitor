import PhotoSwipeLightbox from 'photoswipe/lightbox';
import 'photoswipe/style.css';

export default defineNuxtPlugin({
  parallel: true,
  setup() {
    const openPhotoSwipe = (images, startIndex = 0, options = {}) => {
      const defaultOptions = {
        showHideAnimationType: 'fade',
        bgOpacity: 0.9,
        spacing: 0.1,
        allowPanToNext: true,
        closeOnVerticalDrag: true,
        pinchToClose: true,
        clickToCloseNonZoomable: true,
        zoom: true,
        loop: true,
        wheelToZoom: true,
        arrowKeys: true,
        returnFocus: false,
      };

      const lightbox = new PhotoSwipeLightbox({
        dataSource: images,
        pswpModule: () => import('photoswipe'),
        ...defaultOptions,
        ...options,
      });

      lightbox.init();
      lightbox.loadAndOpen(startIndex);

      lightbox.on('destroy', lightbox.destroy);

      return lightbox;
    };

    return {
      provide: {
        openPhotoSwipe,
      },
    };
  },
});
