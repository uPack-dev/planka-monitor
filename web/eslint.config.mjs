import withNuxt from './.nuxt/eslint.config.mjs';
import prettier from 'eslint-plugin-prettier/recommended';
import tseslint from 'typescript-eslint';

export default withNuxt(
  {
    files: ['**/*.ts'],
    languageOptions: {
      parser: tseslint.parser,
    },
  },
  {
    ignores: ['public/*', 'node_modules/*', '.nuxt/*', 'jsconfig.json'],
    rules: {
      'no-undef': 'off',
      'vue/no-v-html': 'off',
      'vue/prop-name-casing': 'off',
      'vue/multi-word-component-names': 'off',
      '@stylistic/indent': ['error', 2],
    },
  },
  prettier,
);
