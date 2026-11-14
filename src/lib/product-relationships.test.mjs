import assert from "node:assert/strict";
import { afterEach, test } from "node:test";
import {
  getAdminProducts,
  getProductRelationships,
  orderedRelatedProductIds,
  replaceProductRelationships,
  searchAvailableProducts,
  ProductRelationshipApiError,
} from "./product-relationships.ts";

const originalFetch = globalThis.fetch;
afterEach(() => { globalThis.fetch = originalFetch; });

test("loads admin products and relationships, retaining the stored order", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    return Response.json(url.endsWith("/relationships")
      ? [
          { id: "later", relatedProductId: "pig-b", displayOrder: 2 },
          { id: "first", relatedProductId: "pig-a", displayOrder: 0 },
        ]
      : [{ id: "pig-a", title: "Party Pig", slug: "party-pig" }]);
  };
  const [products, relationships] = await Promise.all([
    getAdminProducts(), getProductRelationships("source/id"),
  ]);
  assert.equal(products[0].title, "Party Pig");
  assert.deepEqual(orderedRelatedProductIds(relationships), ["pig-a", "pig-b"]);
  assert.deepEqual(calls.map((call) => call.url), [
    "/api/admin/products",
    "/api/admin/products/source%2Fid/relationships",
  ]);
  assert.ok(calls.every((call) => call.init.cache === "no-store"));
});

test("searches title and slug, excluding the source and selected products", () => {
  const products = [
    { id: "source", title: "Party Pig", slug: "party-pig" },
    { id: "selected", title: "Party Hat", slug: "party-hat" },
    { id: "available", title: "Cloud Pig", slug: "party-cloud" },
    { id: "available-2", title: "Party Banner", slug: "banner" },
  ];
  assert.deepEqual(
    searchAvailableProducts(products, "PARTY", "source", ["selected"]).map((product) => product.id),
    ["available", "available-2"],
  );
  assert.deepEqual(searchAvailableProducts(products, "  ", "source", []), []);
});

test("replacement sends only ordered IDs to the existing admin route", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    return Response.json({ product: { id: "source" }, relationships: [] });
  };
  await replaceProductRelationships("source", ["pig-b", "pig-a"]);
  assert.equal(calls.length, 1);
  assert.equal(calls[0].url, "/api/admin/products/source/relationships");
  assert.equal(calls[0].init.method, "PUT");
  assert.deepEqual(JSON.parse(calls[0].init.body), { relatedProductIds: ["pig-b", "pig-a"] });
});

test("missing related product is reported without exposing server details", async () => {
  globalThis.fetch = async () => Response.json({ message: "secret database detail" }, { status: 404 });
  await assert.rejects(replaceProductRelationships("source", ["missing"]), (error) =>
    error instanceof ProductRelationshipApiError && error.status === 404 &&
    !error.message.includes("secret") && error.message.includes("no longer exists"));
});
