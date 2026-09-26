import type { APIRoute } from "astro";
import { getCatalog } from "@/lib/catalog";
import { getCatalogResults } from "@/lib/catalog-results";

export const GET: APIRoute = async ({ request }) => {
  try {
    const catalog = await getCatalog(request.headers);
    const params = new URL(request.url).searchParams;
    const results = getCatalogResults(catalog.products, params);

    return Response.json(results, {
      headers: { "Cache-Control": "private, no-store" },
    });
  } catch (error) {
    console.error("Unable to load catalog results:", error);
    return Response.json(
      { error: "Unable to load catalog results" },
      { status: 500, headers: { "Cache-Control": "private, no-store" } },
    );
  }
};
