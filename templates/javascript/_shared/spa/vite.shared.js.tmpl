import { loadEnv } from "vite";

// Settings shared by the React and Vue templates: /v1 proxy for dev and preview (same origin, so
// the refresh cookie flows), and a strict CSP <meta> injected into the production build.
export function sharedConfig(mode) {
  const env = loadEnv(mode, process.cwd(), "");
  const proxy = {
    "/v1": { target: env.API_PROXY_TARGET || "http://localhost:3000" },
  };
  return {
    plugins: [cspMeta(env.VITE_API_BASE_URL ?? "")],
    server: { port: 5173, strictPort: true, proxy },
    preview: { port: 4173, strictPort: true, proxy },
  };
}

function cspMeta(apiBase) {
  const apiOrigin = apiBase ? new URL(apiBase).origin : "";
  const csp = [
    "default-src 'self'",
    "script-src 'self'",
    "style-src 'self'",
    "img-src 'self' data:",
    `connect-src 'self'${apiOrigin ? ` ${apiOrigin}` : ""}`,
    "object-src 'none'",
    "base-uri 'self'",
    "form-action 'self'",
  ].join("; ");
  return {
    name: "csp-meta",
    apply: "build",
    transformIndexHtml: () => [
      {
        tag: "meta",
        attrs: { "http-equiv": "Content-Security-Policy", content: csp },
        injectTo: "head-prepend",
      },
    ],
  };
}
