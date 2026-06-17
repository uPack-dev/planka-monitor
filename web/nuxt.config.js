const isDev = process.env.NODE_ENV === 'development';

const gtmID = import.meta.env.GTM_ID;
const turnstileKey = import.meta.env.TURNSTILE_SITE_KEY;

const clientUrl = import.meta.env.NUXT_PUBLIC_CLIENT_URL;
const cmsUrl = import.meta.env.NUXT_PUBLIC_CMS_URL;
const monitorApiUrl = process.env.MONITOR_API_URL || 'http://localhost:8080';

const vitePlugins = [];

const siteName = 'Annonce Mini App';

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@pinia/nuxt',
    '@vueuse/nuxt',
    '@vee-validate/nuxt',
    'nuxt-swiper',
    'nuxt-svgo',
    '@nuxt/eslint',
    '@nuxt/scripts',
    '@nuxtjs/turnstile',
    '@nuxtjs/seo',
    'nuxt-og-image',
  ],

  $development: {
    scripts: {
      registry: {
        googleTagManager: 'mock',
      },
    },
  },

  $production: {
    scripts: {
      registry: {
        googleTagManager: {
          id: gtmID || '',
        },
      },
    },
  },

  components: [
    '@/components',
    { path: '@/components/common', prefix: 'C' },
    { path: '@/components/blocks', prefix: 'B' },
    { path: '@/components/animation', prefix: 'A' },
    { path: '@/components/cards', prefix: 'Card' },
    { path: '@/components/mini', prefix: 'Mini', pathPrefix: false },
    { path: '@/components/sliders', prefix: 'S' },
    { path: '@/components/ui', prefix: 'Ui' },
    { path: '@/components/layout', prefix: 'L' },
  ],
  devtools: { enabled: isDev },
  app: {
    head: {
      layoutTransition: {
        name: 'layout',
        mode: 'out-in',
      },
      viewport:
        'width=device-width, initial-scale=1, maximum-scale=1, minimum-scale=1, user-scalable=no, viewport-fit=cover',
      htmlAttrs: { lang: 'uk' },
      link: [
        { rel: 'preconnect', href: cmsUrl, crossorigin: true },
        { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' },
        {
          rel: 'icon',
          type: 'image/png',
          sizes: '96x96',
          href: '/favicon/favicon-96x96.png',
        },
        {
          rel: 'icon',
          type: 'image/png',
          sizes: '192x192',
          href: '/favicon/web-app-manifest-192x192.png',
        },
        {
          rel: 'icon',
          type: 'image/png',
          sizes: '512x512',
          href: '/favicon/web-app-manifest-512x512.png',
        },
        { rel: 'apple-touch-icon', href: '/favicon/apple-touch-icon.png' },
        {
          rel: 'manifest',
          href: '/favicon/site.webmanifest',
        },
      ],
      script: [
        {
          'data-cfasync': 'false',
          src: 'https://telegram.org/js/telegram-web-app.js',
        },
      ],
      meta: [
        {
          property: 'og:site_name',
          content: siteName,
        },
        {
          name: 'application-name',
          content: siteName,
        },
        {
          property: 'og:type',
          content: 'website',
        },
        {
          property: 'og:locale',
          content: 'uk',
        },
        { charset: 'utf-8' },
        { 'http-equiv': 'Content-Type', content: 'text/html; charset=UTF-8' },
      ],
    },
  },
  css: [
    'reset-css',
    'vue-final-modal/style.css',
    '@/assets/styles/base/index.scss',
    '@fontsource/orelega-one',
    '@fontsource/inter/400.css',
    '@fontsource/inter/500.css',
    '@fontsource/inter/600.css',
    '@fontsource/inter/700.css',
    '@fontsource/ubuntu/700.css',
    '@fontsource/days-one',
  ],

  site: {
    url: clientUrl,
    name: siteName,
  },

  runtimeConfig: {
    serverUrl: import.meta.env.NUXT_SERVER_URL,
    monitorApiUrl,
    public: {
      turnstileKey,
      isDev,
      clientUrl,
      cmsUrl,
      gtmID,
      mockTasks: process.env.NUXT_PUBLIC_MOCK_TASKS || '',
      mockTasksMin: process.env.NUXT_PUBLIC_MOCK_TASKS_MIN || '',
      siteName,
    },
  },

  sourcemap: isDev,
  // swiper: {
  //   modules: [],
  // },
  features: {
    inlineStyles: false,
  },

  experimental: {
    inlineSSRStyles: false,
    // payloadExtraction: true,
    // renderJsonPayloads: true,
    // splitPageChunks: false,
    // sharedPrerenderData: true,
    // asyncContext: true,
  },
  compatibilityDate: '2025-11-03',

  nitro: {
    routeRules: {
      '/**/*.{png,jpeg,jpg,webp,avif,svg,css}': {
        headers: {
          'cache-control':
            'public,max-age=31536000,s-maxage=31536000,immutable',
        },
      },
      '/images/**': {
        headers: {
          'cache-control':
            'public,max-age=31536000,s-maxage=31536000,immutable',
        },
      },
      '/fonts/**': {
        headers: {
          'cache-control':
            'public,max-age=31536000,s-maxage=31536000,immutable',
        },
      },
      '/favicon/**': {
        headers: {
          'cache-control':
            'public,max-age=31536000,s-maxage=31536000,immutable',
        },
      },
      '/_nuxt/**': {
        headers: {
          'cache-control':
            'public,max-age=31536000,s-maxage=31536000,immutable',
          'x-content-type-options': 'nosniff',
        },
      },
    },

    experimental: {
      // tasks: true,
      wasm: false,
    },

    storage: {
      cache: {
        driver: 'lru-cache',
        max: 1000,
        maxSize: 150 * 1024 * 1024,
        ttl: 3600000,
      },
    },
    minify: true,
    compressPublicAssets: {
      gzip: true,
      brotli: true,
    },
    prerender: {
      crawlLinks: false,
      routes: [],
    },
  },

  vite: {
    build: {
      cssCodeSplit: true,
      sourcemap: isDev,
      chunkSizeWarningLimit: 2000,
      assetsInlineLimit: 8192,
      reportCompressedSize: false,
      minify: 'esbuild',
      terserOptions: {
        compress: {
          drop_debugger: !isDev,
          hoist_funs: true,
          passes: 2,
        },
      },
    },
    server: {
      clientPort: 3000,
      hmr: {
        clientPort: 3000,
      },
      allowedHosts: ['webapp.upack-studio.com'],
    },
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: '@use "@/assets/styles/utils" as *;',
          quietDeps: true,
          silenceDeprecations: ['import', 'global-builtin', 'if-function'],
        },
      },
      devSourcemap: isDev,
    },
    optimizeDeps: {
      include: [
        '@fluejs/noscroll',
        '@yeger/vue-masonry-wall',
        'maska/vue',
        'photoswipe',
        'photoswipe/lightbox',
        'qs', // CJS
        'swiper/element/bundle',
        'swiper/vue',
        'typograf',
        'vue-awesome-paginate',
        'vue-final-modal',
      ],
    },
    plugins: vitePlugins,
    // server: {
    //   hmr: {
    //     overlay: false,
    //   },
    //   watch: {
    //     usePolling: true,
    //     interval: 2000,
    //     ignored: ['**/node_modules/**', '**/.git/**'],
    //   },
    // },
  },
  hooks: {
    // https://github.com/nuxt/nuxt/issues/14584#issuecomment-2166544081
    'build:manifest': (manifest) => {
      const updates = {};
      for (const [key, item] of Object.entries(manifest)) {
        const isEntry = key.includes('entry') || key.includes('app');
        const isCriticalCSS =
          key.includes('layout') ||
          key.includes('common') ||
          key.includes('entry');

        updates[key] = {
          ...item,
          dynamicImports: [],
          prefetch: false,
          preload: key.endsWith('.css') && isCriticalCSS,
          modulePreload: isEntry,
        };
      }
      Object.assign(manifest, updates);
    },

    'vite:extendConfig'(config) {
      config.build.rollupOptions.output.manualChunks = function (id) {
        if (id.includes('/layout')) {
          return 'layout';
        }
      };
    },
  },

  eslint: {
    config: {
      stylistic: true, // <---
    },
  },

  ogImage: {
    defaults: {
      extension: 'png',
    },
    runtimeCacheStorage: false,
  },

  robots: {
    blockAiBots: true,
  },

  schemaOrg: {
    defaults: false,
  },

  sitemap: {
    sources: ['/api/__sitemap__/urls'],

    defaults: {
      lastmod: new Date().toISOString(),
    },
  },

  svgo: { defaultImport: 'component', explicitImportsOnly: true },

  turnstile: {
    siteKey: turnstileKey,
  },

  // delayHydration: { debug: import.meta.env.DELAY_HYDRATION_DEBUG, mode: 'mount' },
  veeValidate: {
    autoImports: true,
    componentNames: {
      Form: 'VeeForm',
      Field: 'VeeField',
      FieldArray: 'VeeFieldArray',
      ErrorMessage: 'VeeErrorMessage',
    },
  },
});
