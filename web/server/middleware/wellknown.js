export default defineEventHandler((event) => {
  if (event.req.url === '/.well-known/appspecific/com.chrome.devtools.json') {
    return {};
  }
});
