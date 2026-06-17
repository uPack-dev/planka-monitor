import { ROUTES } from '@/configs/data/routes';
import { defineNuxtRouteMiddleware, useSeoMeta } from 'nuxt/app';

export default defineNuxtRouteMiddleware((to) => {
  const routesList = Object.values(ROUTES);

  const currentRouteConfig = routesList.find(
    (route) =>
      route.link === to.path || route.link === to.path.replace(/\/$/, ''),
  );

  if (currentRouteConfig) {
    useSeoMeta({
      title: currentRouteConfig.seo?.title || currentRouteConfig.title,
      description: currentRouteConfig.seo?.description,
    });
  }

  return true;
});
