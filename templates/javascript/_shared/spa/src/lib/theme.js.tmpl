// Light/dark theme. The choice is the `theme` cookie (functional, no personal data); without it
// the OS preference decides (CSS prefers-color-scheme). public/theme.js applies the cookie before
// the first paint, because the CSP forbids an inline script.

// Only these two values may reach the page: the same check guards theme.js's regex.
export function themeFrom(value) {
  return value === "light" || value === "dark" ? value : undefined;
}

// What the page shows now: the explicit choice, else the OS preference.
export function currentTheme() {
  const chosen = themeFrom(document.documentElement.dataset.theme);
  if (chosen) return chosen;
  return matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

// Applies the choice now and keeps it for a year.
export function setTheme(theme) {
  document.documentElement.dataset.theme = theme;
  const secure = location.protocol === "https:" ? "; Secure" : "";
  // biome-ignore lint/suspicious/noDocumentCookie: a functional preference that JS must write; the Cookie Store API is not in every browser
  document.cookie = `theme=${theme}; Path=/; Max-Age=31536000; SameSite=Lax${secure}`;
}
