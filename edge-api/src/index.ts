export interface Env {
  APP_NAME: string;
  COURSE_NAME: string;
  API_TOKEN: string;
  ADMIN_EMAIL: string;
  SETTINGS: KVNamespace;
}

function json(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data, null, 2), {
    status,
    headers: { "content-type": "application/json; charset=UTF-8" },
  });
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const path = url.pathname;

    console.log("request", {
      method: request.method,
      path,
      colo: request.cf?.colo,
      country: request.cf?.country,
    });

    if (path === "/health") {
      return json({
        status: "ok",
        service: env.APP_NAME ?? "edge-api",
        timestamp: new Date().toISOString(),
      });
    }

    if (path === "/") {
      return json({
        app: env.APP_NAME ?? "edge-api",
        course: env.COURSE_NAME ?? "devops-core",
        message: "Hello from Cloudflare Workers v3",
        timestamp: new Date().toISOString(),
        routes: ["/", "/health", "/edge", "/config", "/secret-check", "/counter"],
      });
    }

    if (path === "/edge") {
      return json({
        colo: request.cf?.colo ?? null,
        country: request.cf?.country ?? null,
        city: request.cf?.city ?? null,
        asn: request.cf?.asn ?? null,
        httpProtocol: request.cf?.httpProtocol ?? null,
        tlsVersion: request.cf?.tlsVersion ?? null,
        timestamp: new Date().toISOString(),
      });
    }

    if (path === "/config") {
      return json({
        appName: env.APP_NAME ?? null,
        courseName: env.COURSE_NAME ?? null,
        note: "Plaintext vars are fine for non-sensitive config only.",
      });
    }

    if (path === "/secret-check") {
      return json({
        apiTokenConfigured: Boolean(env.API_TOKEN),
        adminEmailConfigured: Boolean(env.ADMIN_EMAIL),
        note: "Secret values are intentionally not returned.",
      });
    }

    if (path === "/counter") {
      const current = Number((await env.SETTINGS.get("visits")) ?? "0");
      const next = current + 1;
      await env.SETTINGS.put("visits", String(next));
      return json({
        visits: next,
        storedKey: "visits",
        timestamp: new Date().toISOString(),
      });
    }

    return json({ error: "Not Found", path }, 404);
  },
};