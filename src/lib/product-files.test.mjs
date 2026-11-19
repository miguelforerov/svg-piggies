import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { afterEach, test } from "node:test";
import formatConfig from "../../config/product-file-formats.json" with { type: "json" };
import {
  ProductFileApiError, acceptedProductFileExtensions, deleteProductFile, getProductFiles,
  supportedProductFileLabel, uploadProductFile, validateProductFile,
} from "./product-files.ts";

const originalFetch = globalThis.fetch;
afterEach(() => { globalThis.fetch = originalFetch; });

const file = (name = "pig.svg", type = "image/svg+xml") => new File(["<svg></svg>"], name, { type });
const authorization = (token = "token") => ({
  uploadToken: token,
  uploadUrl: "https://r2.example.test/files/upload",
  headers: { "Content-Type": ["image/svg+xml"], "If-None-Match": ["*"], "x-amz-meta-example": ["one"] },
  expiresAt: new Date(Date.now() + 60_000).toISOString(),
});
const record = { id: "file-id", productId: "product-id", fileName: "pig.svg", displayName: "pig.svg",
  format: "svg", contentType: "image/svg+xml", fileSizeBytes: 11, availabilityStatus: "active",
  displayOrder: 0, createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" };

test("shared JSON defines the five format identifiers, extensions, and MIME types", () => {
  assert.equal(typeof formatConfig._comment, "string");
  assert.deepEqual(formatConfig.formats, [
    { format: "svg", extension: ".svg", contentTypes: ["image/svg+xml"] },
    { format: "png", extension: ".png", contentTypes: ["image/png"] },
    { format: "dxf", extension: ".dxf", contentTypes: ["image/vnd.dxf", "application/dxf"] },
    { format: "pdf", extension: ".pdf", contentTypes: ["application/pdf"] },
    { format: "zip", extension: ".zip", contentTypes: ["application/zip", "application/x-zip-compressed", "application/octet-stream"] },
  ]);
  assert.equal(acceptedProductFileExtensions, formatConfig.formats.map(({ extension }) => extension).join(","));
  assert.equal(supportedProductFileLabel, "SVG, PNG, DXF, PDF, or ZIP");
  const generatedType = `// Code generated from config/product-file-formats.json; DO NOT EDIT.\nexport type ProductFileFormat = ${formatConfig.formats.map(({ format }) => JSON.stringify(format)).join(" | ")};\n`;
  assert.equal(readFileSync(new URL("./product-file-formats.generated.ts", import.meta.url), "utf8"), generatedType);
});

test("accepts supported extensions and MIME types, including ZIP variants", () => {
  for (const [name, type] of [
    ["pig.svg", "image/svg+xml"], ["pig.png", "image/png"],
    ["pig.dxf", "image/vnd.dxf"], ["pig.dxf", "application/dxf"],
    ["pig.pdf", "application/pdf"], ["PIG.SVG", "image/svg+xml"],
    ["pig.dxf", ""],
    ["bundle.zip", "application/zip"], ["bundle.zip", "application/x-zip-compressed"],
    ["bundle.zip", "application/octet-stream"], ["bundle.zip", ""],
  ]) assert.equal(validateProductFile(file(name, type)), null, `${name} ${type}`);
  for (const [name, type] of [
    ["pig.exe", "application/octet-stream"], ["bundle.zip", "application/pdf"],
    ["pig.svg", "application/pdf"], ["pig.pdf", "image/png"],
    ["pig.dxf", "application/octet-stream"], ["svg", "image/svg+xml"],
  ]) assert.ok(validateProductFile(file(name, type)), `${name} ${type}`);
  assert.match(validateProductFile(file("pig.exe", "application/octet-stream")), /SVG, PNG, DXF, PDF, or ZIP/);
});

test("enforces the exact byte-size and filename boundaries before a network request", async () => {
  for (const size of [1, 50_000_000]) {
    assert.equal(validateProductFile({ name: "pig.svg", type: "image/svg+xml", size }), null);
  }
  for (const size of [0, 50_000_001]) {
    assert.match(validateProductFile({ name: "pig.svg", type: "image/svg+xml", size }), /size/);
  }
  for (const size of [1, 50_000_000]) {
    assert.equal(validateProductFile({ name: "bundle.zip", type: "application/zip", size }), null);
  }
  for (const size of [0, 50_000_001]) {
    assert.match(validateProductFile({ name: "bundle.zip", type: "application/zip", size }), /size/);
  }
  for (const name of [" pig.svg", "pig.svg ", "bad/name.svg", "bad\\name.svg", "bad\nname.svg", "🐷".repeat(64) + ".svg"]) {
    assert.match(validateProductFile({ name, type: "image/svg+xml", size: 1 }), /Filename/);
  }
  globalThis.fetch = () => { throw new Error("fetch must not run"); };
  await assert.rejects(uploadProductFile("product-id", file("pig.exe", "application/octet-stream"), () => {}), /SVG, PNG, DXF, PDF, or ZIP/);
});

test("ZIP authorization uses its format and canonical MIME when the browser MIME is empty", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    if (calls.length === 1) return Response.json({ ...authorization(), headers: { "Content-Type": ["application/zip"], "If-None-Match": ["*"] } }, { status: 201 });
    if (calls.length === 2) return new Response(null, { status: 200 });
    return Response.json({ ...record, fileName: "bundle.zip", format: "zip", contentType: "application/zip" }, { status: 201 });
  };
  const archive = file("bundle.zip", "");
  const confirmed = await uploadProductFile("product", archive, () => {});
  assert.deepEqual(JSON.parse(calls[0].init.body), { fileName: "bundle.zip", format: "zip", contentType: "application/zip", fileSizeBytes: archive.size });
  assert.equal(calls[1].init.body, archive);
  assert.equal(calls[1].init.headers.get("Content-Type"), "application/zip");
  assert.equal(confirmed.format, "zip");
});

test("authorizes, PUTs original bytes with every signed header, then confirms", async () => {
  const calls = [];
  const stages = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    if (calls.length === 1) return Response.json(authorization(), { status: 201 });
    if (calls.length === 2) return new Response(null, { status: 200 });
    return Response.json(record, { status: 201 });
  };
  const original = file();
  assert.deepEqual(await uploadProductFile("product/id", original, (stage) => stages.push(stage)), record);
  assert.deepEqual(stages, ["authorizing", "uploading", "confirming"]);
  assert.equal(calls.length, 3);
  assert.equal(calls[0].url, "/api/admin/products/product%2Fid/files/upload-authorization");
  assert.deepEqual(JSON.parse(calls[0].init.body), { fileName: original.name, format: "svg", contentType: "image/svg+xml", fileSizeBytes: original.size });
  assert.equal(calls[1].url, authorization().uploadUrl);
  assert.equal(calls[1].init.method, "PUT");
  assert.equal(calls[1].init.body, original);
  assert.equal(calls[1].init.headers.get("Content-Type"), "image/svg+xml");
  assert.equal(calls[1].init.headers.get("If-None-Match"), "*");
  assert.equal(calls[1].init.headers.get("x-amz-meta-example"), "one");
  assert.equal(calls[1].init.credentials, "omit");
  assert.equal(calls[1].init.redirect, "error");
  assert.equal(calls[2].url, "/api/admin/products/product%2Fid/files");
  assert.deepEqual(JSON.parse(calls[2].init.body), { uploadToken: "token" });
});

test("an empty browser MIME type uses the documented format for authorization", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    if (calls.length === 1) return Response.json({ ...authorization(), headers: { "Content-Type": ["image/vnd.dxf"], "If-None-Match": ["*"] } }, { status: 201 });
    if (calls.length === 2) return new Response(null, { status: 200 });
    return Response.json({ ...record, format: "dxf" }, { status: 201 });
  };
  await uploadProductFile("product", file("pig.dxf", ""), () => {});
  assert.deepEqual(JSON.parse(calls[0].init.body), { fileName: "pig.dxf", format: "dxf", contentType: "image/vnd.dxf", fileSizeBytes: 11 });
});

test("a failed PUT never confirms; a retry requests a new authorization", async () => {
  const calls = [];
  globalThis.fetch = async (url) => {
    calls.push(url);
    if (calls.length === 1) return Response.json(authorization("old"), { status: 201 });
    if (calls.length === 2) return new Response(null, { status: 403 });
    if (calls.length === 3) return Response.json(authorization("new"), { status: 201 });
    if (calls.length === 4) return new Response(null, { status: 200 });
    return Response.json(record, { status: 201 });
  };
  await assert.rejects(uploadProductFile("product", file(), () => {}), /Direct file upload failed/);
  assert.equal(calls.length, 2);
  await uploadProductFile("product", file(), () => {});
  assert.equal(calls.length, 5);
  assert.equal(calls[0], calls[2]);
});

test("expired or malformed authorization stops before PUT and exposes no URL", async () => {
  for (const invalid of [
    { ...authorization(), expiresAt: new Date(Date.now() - 1000).toISOString() },
    { ...authorization(), expiresAt: "invalid" },
    { ...authorization(), headers: { "Content-Type": ["image/svg+xml"] } },
    { ...authorization(), headers: { "Content-Type": ["application/pdf"], "If-None-Match": ["*"] } },
  ]) {
    let calls = 0;
    globalThis.fetch = async () => { calls++; return Response.json(invalid, { status: 201 }); };
    await assert.rejects(uploadProductFile("product", file(), () => {}), (error) =>
      error instanceof ProductFileApiError && !error.message.includes(invalid.uploadUrl));
    assert.equal(calls, 1);
  }
});

test("expiry after PUT prevents a new confirmation", async () => {
  const calls = [];
  globalThis.fetch = async (url) => {
    calls.push(url);
    if (calls.length === 1) return Response.json({ ...authorization(), expiresAt: new Date(Date.now() + 1000).toISOString() }, { status: 201 });
    await new Promise((resolve) => setTimeout(resolve, 1100));
    return new Response(null, { status: 200 });
  };
  await assert.rejects(uploadProductFile("product", file(), () => {}), (error) => error instanceof ProductFileApiError && error.status === 410);
  assert.equal(calls.length, 2);
});

test("list and delete use only public IDs; empty list and 204 work", async () => {
  const calls = [];
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    return calls.length === 1 ? Response.json([], { status: 200 }) : new Response(null, { status: 204 });
  };
  assert.deepEqual(await getProductFiles("product/id"), []);
  await deleteProductFile("product/id", "file/id");
  assert.equal(calls[0].url, "/api/admin/products/product%2Fid/files");
  assert.equal(calls[1].url, "/api/admin/products/product%2Fid/files/file%2Fid");
  assert.equal(calls[1].init.method, "DELETE");
  assert.equal(calls[1].init.body, undefined);
  assert.equal(calls[1].init.cache, "no-store");
});

test("API status and network errors use safe messages without server details", async () => {
  for (const status of [400, 401, 403, 404, 409, 410, 502, 503, 500]) {
    globalThis.fetch = async () => Response.json({ message: "secret R2 object key" }, { status });
    await assert.rejects(getProductFiles("product"), (error) => error instanceof ProductFileApiError &&
      error.status === status && !error.message.includes("secret"));
  }
  globalThis.fetch = async () => { throw new Error("https://r2.example.test/private-token"); };
  await assert.rejects(getProductFiles("product"), (error) => error instanceof ProductFileApiError && !error.message.includes("private-token"));
  globalThis.fetch = async () => Response.json({ message: "secret" }, { status: 502 });
  await assert.rejects(deleteProductFile("product", "file"), (error) => error instanceof ProductFileApiError && !error.message.includes("secret"));
});
