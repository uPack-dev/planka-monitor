import { ROUTES } from '@/configs/data/routes';
import contacts from '@/configs/data/contacts';
import { defineStore } from 'pinia';

const footerRoutes = Object.fromEntries(
  Object.entries(ROUTES).filter(([routeName]) => routeName !== 'MAIN'),
);

export const useGlobalStore = defineStore('global', () => {
  const menu = {
    header: Object.values(ROUTES),
    footer: Object.values(footerRoutes),
  };

  return {
    menu,
    contacts,
  };
});
