import { afterEach, expect, test, vi } from "vitest";

const ok = (data, status = 200) =>
  new Response(JSON.stringify({ meta: { success: status < 400 }, data }), {
    status,
  });

afterEach(() => {
  vi.unstubAllGlobals();
  vi.resetModules();
});

test("login keeps the access token in memory and sends the CSRF header", async () => {
  const fetchMock = vi.fn(async (url, init) => {
    if (url.endsWith("/v1/auth/login"))
      return ok({ access_token: "a1", expires_in: 900 });
    return ok({
      authorization: new Headers(init.headers).get("authorization"),
    });
  });
  vi.stubGlobal("fetch", fetchMock);
  const auth = await import("./auth.js");
  await auth.login("a@example.com", "password123");
  const [, loginInit] = fetchMock.mock.calls[0] ?? [];
  expect(loginInit?.credentials).toBe("include");
  expect(new Headers(loginInit?.headers).get("x-csrf-protection")).toBe("1");
  expect(await auth.api("/v1/users/me")).toEqual({
    authorization: "Bearer a1",
  });
  const [, apiInit] = fetchMock.mock.calls[1] ?? [];
  expect(apiInit?.credentials).toBe("omit");
});

test("a 401 refreshes once and retries", async () => {
  let n = 0;
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url, init) => {
      if (url.endsWith("/v1/auth/login")) return ok({ access_token: "old" });
      if (url.endsWith("/v1/auth/refresh"))
        return ok({ access_token: `new${++n}` });
      const auth = new Headers(init.headers).get("authorization");
      return auth === "Bearer new1" ? ok({ id: 1 }) : ok(null, 401);
    }),
  );
  const auth = await import("./auth.js");
  await auth.login("a@example.com", "password123");
  expect(await auth.api("/v1/users/me")).toEqual({ id: 1 });
  expect(n).toBe(1);
});

test("concurrent callers share one refresh", async () => {
  let refreshes = 0;
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url) => {
      if (url.endsWith("/v1/auth/refresh")) {
        refreshes++;
        await new Promise((r) => setTimeout(r, 20));
        return ok({ access_token: "t" });
      }
      return ok({});
    }),
  );
  const auth = await import("./auth.js");
  await Promise.all([auth.api("/v1/a"), auth.api("/v1/b"), auth.api("/v1/c")]);
  expect(refreshes).toBe(1);
});

test("no session: the API error surfaces as 401", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => ok(null, 401)),
  );
  const auth = await import("./auth.js");
  await expect(auth.api("/v1/users/me")).rejects.toMatchObject({ status: 401 });
});

test("a refresh still in flight cannot restore the token after logout", async () => {
  let release = () => {};
  let started = () => {};
  const refreshStarted = new Promise((r) => {
    started = r;
  });
  const calls = [];
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url) => {
      calls.push(new URL(url, "http://x").pathname);
      if (url.endsWith("/v1/auth/refresh")) {
        if (calls.length > 1) return ok(null, 401); // after logout: no session
        started();
        await new Promise((r) => {
          release = r;
        });
        return ok({ access_token: "late" });
      }
      return ok(null);
    }),
  );
  const auth = await import("./auth.js");
  const pending = auth.refresh();
  await refreshStarted;
  const out = auth.logout();
  release();
  expect(await pending).toBe(false);
  await out;
  // logout waited for the refresh (its rotated cookie is the one revoked) and nothing restored it
  expect(calls).toEqual(["/v1/auth/refresh", "/v1/auth/logout"]);
  await expect(auth.api("/v1/users/me")).rejects.toMatchObject({ status: 401 });
});

test("logout fails loudly when the API cannot revoke the session", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => ok(null, 503)),
  );
  const auth = await import("./auth.js");
  await expect(auth.logout()).rejects.toMatchObject({ status: 503 });
});

test("refreshes are serialized across tabs with the Web Locks API", async () => {
  const request = vi.fn((_name, fn) => fn());
  vi.stubGlobal("navigator", { locks: { request } });
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => ok({ access_token: "t" })),
  );
  const auth = await import("./auth.js");
  expect(await auth.refresh()).toBe(true);
  expect(request).toHaveBeenCalledWith("auth-refresh", expect.any(Function));
});
