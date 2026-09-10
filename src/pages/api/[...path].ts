import { env } from "cloudflare:workers";
import type { APIRoute } from "astro";

export const ALL: APIRoute = async ({ request }) => {
  const api = env.API;

  if (!api) {
    return new Response("API service binding is not configured", {
      status: 500,
    });
  }

  const url = new URL(request.url);
  const targetUrl = new URL(url);

  targetUrl.hostname = "svg-piggies-api";
  targetUrl.port = "";
  targetUrl.protocol = "https:";

  const targetRequest = new Request(targetUrl, request);

  return api.fetch(targetRequest);
};