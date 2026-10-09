// Stand-in for the backend API in AUTH_COOKIE_MODE so the browser flow can be tested without
// Go/Node API, database or Redis. Same contract: refresh token in an HttpOnly cookie scoped to
// /v1/auth, access token in the JSON body, X-CSRF-Protection required on auth calls.
import { createServer } from "node:http";

const USER = { id: "0199a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b", name: "Test Admin", email: "admin@example.com", roles: [{ id: "0199a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a01", code: "admin", name: "Admin" }] };
const COOKIE = "__Secure-refresh_token";
let seq = 0;
let currentRefresh = "";
const access = new Set();

const send = (res, status, data, message = "OK", cookie) => {
  const headers = { "content-type": "application/json" };
  if (cookie !== undefined) {
    headers["set-cookie"] = `${COOKIE}=${cookie}; Path=/v1/auth; Max-Age=${cookie ? 3600 : 0}; HttpOnly; Secure; SameSite=Strict`;
  }
  res.writeHead(status, headers);
  res.end(JSON.stringify({ meta: { success: status < 400, message }, data }));
};
const issue = (res) => {
  seq++;
  currentRefresh = `refresh-${seq}`;
  access.add(`access-${seq}`);
  send(res, 200, { access_token: `access-${seq}`, expires_in: 900 }, "OK", currentRefresh);
};

createServer(async (req, res) => {
  let raw = "";
  for await (const chunk of req) raw += chunk;
  const body = raw ? JSON.parse(raw) : {};
  const cookie = /__Secure-refresh_token=([^;]+)/.exec(req.headers.cookie ?? "")?.[1];
  if (req.url?.startsWith("/v1/auth/") && req.headers["x-csrf-protection"] !== "1") {
    return send(res, 403, null, "csrf check failed");
  }
  if (req.url === "/v1/auth/login") {
    return body.email === USER.email && body.password === "password123"
      ? issue(res)
      : send(res, 401, null, "invalid credentials");
  }
  if (req.url === "/v1/auth/refresh") {
    return cookie && cookie === currentRefresh ? issue(res) : send(res, 401, null, "session expired");
  }
  if (req.url === "/v1/auth/logout") {
    currentRefresh = "";
    return send(res, 200, null, "OK", "");
  }
  if (req.url === "/v1/users/me") {
    const token = (req.headers.authorization ?? "").replace("Bearer ", "");
    return access.has(token) ? send(res, 200, USER) : send(res, 401, null, "unauthorized");
  }
  send(res, 404, null, "route not found");
}).listen(Number(process.env.MOCK_API_PORT ?? 3998));
