import assert from "node:assert/strict";
import { afterEach, test } from "node:test";
import { uploadProductImage, validateImage, getProductImages, deleteProductImage, reorderProductImages, setPrimaryProductImage, updateProductImageAltText, ProductImageApiError } from "./product-images.ts";

const originalFetch = globalThis.fetch;
afterEach(() => { globalThis.fetch = originalFetch; });
const file = () => new File(["image bytes"], "pig.webp", { type: "image/webp" });
const auth = (token = "token") => ({ uploadToken: token, uploadUrl: "https://r2.example.test/upload", headers: { "Content-Type": ["image/webp"], "x-amz-meta-test": ["one"] }, expiresAt: new Date(Date.now() + 60_000).toISOString() });

test("validates MIME, size, and filename before network calls", async () => {
  assert.equal(validateImage(file()), null);
  assert.match(validateImage(new File(["x"], "pig.gif", { type: "image/gif" })), /JPEG/);
  assert.match(validateImage(new File([], "pig.webp", { type: "image/webp" })), /size/);
  assert.match(validateImage(new File([new Uint8Array(10_485_761)], "pig.webp", { type: "image/webp" })), /size/);
  assert.match(validateImage(new File(["x"], "a".repeat(256) + ".webp", { type: "image/webp" })), /Filename/);
  globalThis.fetch = () => { throw new Error("fetch must not run"); };
  await assert.rejects(uploadProductImage("id", new File([], "pig.webp", { type: "image/webp" }), "", () => {}), /size/);
});

test("authorizes, PUTs original bytes with signed headers, then confirms", async () => {
  const calls = [];
  const stages = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    if (calls.length === 1) return Response.json(auth(), { status: 201 });
    if (calls.length === 2) return new Response(null, { status: 200, headers: { ETag: '"etag"' } });
    return Response.json({ id: "image-id" }, { status: 201 });
  };
  const image = file();
  await uploadProductImage("product-id", image, "Pig image", (stage) => stages.push(stage), { isPrimary: true, displayOrder: 0 });
  assert.deepEqual(stages, ["authorizing", "uploading", "confirming"]);
  assert.equal(calls[0].url, "/api/admin/products/product-id/images/uploads");
  assert.deepEqual(JSON.parse(calls[0].init.body), { originalFilename: "pig.webp", contentType: "image/webp", fileSizeBytes: image.size });
  assert.equal(calls[1].url, "https://r2.example.test/upload");
  assert.equal(calls[1].init.method, "PUT");
  assert.equal(calls[1].init.body, image);
  assert.equal(calls[1].init.headers.get("Content-Type"), "image/webp");
  assert.equal(calls[1].init.headers.get("x-amz-meta-test"), "one");
  assert.deepEqual(JSON.parse(calls[2].init.body), { uploadToken: "token", altText: "Pig image", isPrimary: true, displayOrder: 0 });
});

test("never confirms a failed PUT; retry obtains a new authorization", async () => {
  const calls = [];
  globalThis.fetch = async (url) => {
    calls.push(url);
    if (calls.length === 1) return Response.json(auth("old"), { status: 201 });
    if (calls.length === 2) return new Response(null, { status: 403 });
    if (calls.length === 3) return Response.json(auth("new"), { status: 201 });
    if (calls.length === 4) return new Response(null, { status: 200 });
    return Response.json({ id: "image-id" }, { status: 201 });
  };
  await assert.rejects(uploadProductImage("id", file(), "", () => {}), /Direct image upload failed/);
  assert.equal(calls.length, 2);
  await uploadProductImage("id", file(), "", () => {});
  assert.equal(calls.length, 5);
  assert.equal(calls[0], calls[2]);
});

test("expired authorization and API errors remain safe", async () => {
  globalThis.fetch = async () => Response.json({ ...auth(), expiresAt: new Date(Date.now() - 1000).toISOString() }, { status: 201 });
  await assert.rejects(uploadProductImage("id", file(), "", () => {}), (error) => error instanceof ProductImageApiError && error.status === 410);
  globalThis.fetch = async () => Response.json({ message: "secret signed URL" }, { status: 502 });
  await assert.rejects(getProductImages("id"), (error) => error instanceof ProductImageApiError && !error.message.includes("secret"));
  let calls = 0;
  globalThis.fetch = async () => {
    calls += 1;
    if (calls === 1) return Response.json(auth(), { status: 201 });
    throw new Error("https://r2.example.test/secret-token");
  };
  await assert.rejects(uploadProductImage("id", file(), "", () => {}), (error) => !error.message.includes("secret-token"));
});

test("DELETE uses only product and image IDs and accepts an empty 204 response", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => { calls.push({ url, init }); return new Response(null, { status: 204 }); };
  await deleteProductImage("product/id", "image/id");
  assert.equal(calls.length, 1);
  assert.equal(calls[0].url, "/api/admin/products/product%2Fid/images/image%2Fid");
  assert.equal(calls[0].init.method, "DELETE");
  assert.equal(calls[0].init.body, undefined);
  assert.equal(calls[0].init.cache, "no-store");
});

test("DELETE surfaces safe errors without parsing server details", async () => {
  globalThis.fetch = async () => Response.json({ message: "secret R2 object key" }, { status: 502 });
  await assert.rejects(deleteProductImage("product", "image"), (error) => error instanceof ProductImageApiError && error.status === 502 && !error.message.includes("secret"));
  globalThis.fetch = async () => new Response(null, { status: 202 });
  await assert.rejects(deleteProductImage("product", "image"), (error) => error instanceof ProductImageApiError && error.status === 202);
});

test("setting a primary image sends only IDs and waits for 204", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => { calls.push({ url, init }); return new Response(null, { status: 204 }); };
  await setPrimaryProductImage("product/id", "image/id");
  assert.equal(calls.length, 1);
  assert.equal(calls[0].url, "/api/admin/products/product%2Fid/images/image%2Fid/primary");
  assert.equal(calls[0].init.method, "PUT");
  assert.equal(calls[0].init.body, undefined);
  assert.equal(calls[0].init.cache, "no-store");
  globalThis.fetch = async () => Response.json({ message: "secret database detail" }, { status: 500 });
  await assert.rejects(setPrimaryProductImage("product", "image"), (error) => error instanceof ProductImageApiError && error.status === 500 && !error.message.includes("secret"));
});

test("alt text PATCH preserves omitted, null, empty, and text payloads", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => { calls.push({ url, init }); return new Response(null, { status: 204 }); };
  await updateProductImageAltText("product", "image");
  await updateProductImageAltText("product", "image", null);
  await updateProductImageAltText("product", "image", "");
  await updateProductImageAltText("product", "image", "Pink pig");
  assert.equal(calls.length, 4);
  assert.ok(calls.every((call) => call.url === "/api/admin/products/product/images/image" && call.init.method === "PATCH" && call.init.cache === "no-store"));
  assert.deepEqual(calls.map((call) => JSON.parse(call.init.body)), [{}, { altText: null }, { altText: "" }, { altText: "Pink pig" }]);
  globalThis.fetch = async () => Response.json({ message: "secret SQL detail" }, { status: 500 });
  await assert.rejects(updateProductImageAltText("product", "image", "New"), (error) => error instanceof ProductImageApiError && error.status === 500 && !error.message.includes("secret"));
});

test("reorder sends only the complete image IDs and requires 204", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => { calls.push({ url, init }); return new Response(null, { status: 204 }); };
  await reorderProductImages("product/id", ["second", "first"]);
  assert.equal(calls[0].url, "/api/admin/products/product%2Fid/images/order");
  assert.equal(calls[0].init.method, "PUT");
  assert.deepEqual(JSON.parse(calls[0].init.body), { imageIds: ["second", "first"] });
  assert.equal(calls[0].init.cache, "no-store");
  globalThis.fetch = async () => Response.json({ message: "secret SQL detail" }, { status: 500 });
  await assert.rejects(reorderProductImages("product", []), (error) => error instanceof ProductImageApiError && !error.message.includes("secret"));
});
