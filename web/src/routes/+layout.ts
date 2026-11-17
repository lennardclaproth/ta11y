// The app is a client-only SPA: the session lives in a cookie on the API host and every page
// is guarded in the browser, so nothing can be rendered or prerendered ahead of time.
export const ssr = false;
export const prerender = false;
