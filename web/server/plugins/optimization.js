// style preloads
const stylesheetLinkRegexp = /<link rel="stylesheet" href="(?<href>[^"]+)">/gm;

function getPreloadLink(href) {
  return `<link rel="preload" as="style" href="${href}">`;
}

// script preloads
const scriptPreloadRegexp =
  /<link rel="modulepreload" as="script" crossorigin href="(?<href>[^"]+)">\n?/gm;

// any prefetches
const prefetchRegexp =
  /<link rel="prefetch" as="((script)|(style))" (crossorigin )?href="(?<href>[^"]+)">\n?/gm;

const jsScriptRegexp =
  /<script(?![^>]*\bdata-cfasync=)(?![^>]*type="application\/json")/gm;

const devAuthBootstrapScript = `<script data-cfasync="false">
(function () {
  var cookieName = 'monitor_dev_auth';
  var storageName = 'monitorDevAuthToken';
  var bootstrapName = '__MONITOR_DEV_AUTH_TOKEN__';
  var paramNames = ['monitorDevToken', 'monitorDevAuth', 'monitor_dev_auth', 'devAuth'];
  var expiredCookie = [
    cookieName + '=',
    'path=/',
    'max-age=0',
    'expires=Thu, 01 Jan 1970 00:00:00 GMT',
    'SameSite=Lax'
  ].join('; ');
  var url = new URL(window.location.href);
  var token = '';

  try {
    window.localStorage.removeItem(storageName);
  } catch (error) {}
  document.cookie = expiredCookie;

  for (var index = 0; index < paramNames.length; index += 1) {
    token = url.searchParams.get(paramNames[index]) || '';
    if (token) break;
  }

  if (!token) return;

  token = token.trim();
  if (!token) return;

  window[bootstrapName] = token;

  for (var removeIndex = 0; removeIndex < paramNames.length; removeIndex += 1) {
    url.searchParams.delete(paramNames[removeIndex]);
  }

  window.history.replaceState({}, '', url.pathname + url.search + url.hash);
})();
</script>`;

function disableCloudflareRocketLoader(content) {
  return content.replace(jsScriptRegexp, '<script data-cfasync="false"');
}

function disableCloudflareRocketLoaderForHTML(html) {
  for (const key of ['head', 'body', 'bodyPrepend', 'bodyAppend']) {
    if (Array.isArray(html[key])) {
      html[key] = html[key].map(disableCloudflareRocketLoader);
    }
  }
}

export default defineNitroPlugin((nitroApp) => {
  const {
    public: { isDev },
  } = useRuntimeConfig();

  if (isDev) return;

  nitroApp.hooks.hook('render:html', (html) => {
    const [headContent] = html.head;
    let headContentLocal = headContent;

    // add style preloads
    const stylesheetMatches = headContent.matchAll(stylesheetLinkRegexp);

    for (const stylesheetMatch of stylesheetMatches) {
      const { href } = stylesheetMatch.groups;
      const preloadLink = getPreloadLink(href);

      headContentLocal = `${preloadLink}${headContentLocal}`;
    }

    // remove preload scripts
    headContentLocal = headContentLocal.replace(scriptPreloadRegexp, '');

    // remove prefetches
    headContentLocal = headContentLocal.replace(prefetchRegexp, '');

    html.head[0] = devAuthBootstrapScript + headContentLocal;
    disableCloudflareRocketLoaderForHTML(html);
  });
});
