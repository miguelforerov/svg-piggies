import assert from "node:assert/strict";
import { test } from "node:test";
import { createRequire } from "node:module";
import { build } from "esbuild";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";

const require = createRequire(import.meta.url);

async function loadModule(entry) {
  const result = await build({ entryPoints: [entry], bundle: true, write: false,
    platform: "node", format: "cjs", tsconfig: "tsconfig.json", jsx: "automatic",
    external: ["react", "react-dom", "react-icons/*"] });
  const module = { exports: {} };
  new Function("module", "exports", "require", result.outputFiles[0].text)(module, module.exports, require);
  return module.exports;
}

test("store card uses the primary image alt text and placeholder when absent", async () => {
  const { default: PhotoCard } = await loadModule("src/components/store/ui/PhotoCard.tsx");
  const image = renderToStaticMarkup(React.createElement(PhotoCard, { image: { id: "image", url: "https://signed.test/pig", altText: "Pink pig" }, title: "Pig", slug: "pig" }));
  assert.match(image, /https:\/\/signed\.test\/pig/);
  assert.match(image, /alt="Pink pig"/);
  const empty = renderToStaticMarkup(React.createElement(PhotoCard, { title: "Pig", slug: "pig" }));
  assert.match(empty, /product-placeholder\.jpg/);
  assert.match(empty, /alt="Pig"/);
  const explicitEmpty = renderToStaticMarkup(React.createElement(PhotoCard, { image: { id: "image", url: "https://signed.test/pig", altText: "" }, title: "Pig", slug: "pig" }));
  assert.match(explicitEmpty, /alt=""/);
  const unset = renderToStaticMarkup(React.createElement(PhotoCard, { image: { id: "image", url: "https://signed.test/pig", altText: null }, title: "Pig", slug: "pig" }));
  assert.match(unset, /alt="Pig"/);
});

test("product detail gallery renders every image and its alt text", async () => {
  const { default: ProductGallery } = await loadModule("src/components/store/product/ProductGallery.tsx");
  const markup = renderToStaticMarkup(React.createElement(ProductGallery, { slug: "pig", images: [
    { url: "https://signed.test/one", altText: "Front", width: 800, height: 800 },
    { url: "https://signed.test/two", altText: "Back", width: 800, height: 800 },
  ] }));
  assert.match(markup, /alt="Front"/);
  assert.match(markup, /alt="Back"/);
  const cleared = renderToStaticMarkup(React.createElement(ProductGallery, { slug: "pig", images: [{ id: "image", url: "https://signed.test/one", altText: "", width: 800, height: 800 }] }));
  assert.match(cleared, /alt=""/);
});

test("expired image URLs are refreshed through the public endpoint without cache", async () => {
  const { refreshPublicProductImages } = await loadModule("src/lib/public-product-images.ts");
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async (path, init) => {
      assert.equal(path, "/api/products/pig/images");
      assert.equal(init.cache, "no-store");
      return Response.json([{ id: "image", imageUrl: "https://signed.test/new", displayOrder: 0, isPrimary: true }]);
    };
    const result = await refreshPublicProductImages("pig");
    assert.equal(result[0].url, "https://signed.test/new");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("catalog card selects primary while gallery follows display order", async () => {
  const { normalizePublicImages } = await loadModule("src/lib/product-image-mapping.ts");
  const { default: PhotoCard } = await loadModule("src/components/store/ui/PhotoCard.tsx");
  const { default: ProductGallery } = await loadModule("src/components/store/product/ProductGallery.tsx");
  const images = normalizePublicImages([
    { id: "first", imageUrl: "https://signed.test/first", displayOrder: 0, isPrimary: false, altText: "First" },
    { id: "selected", imageUrl: "https://signed.test/selected", displayOrder: 1, isPrimary: true, altText: "Selected" },
  ]);
  assert.equal(images[0].id, "first");
  const card = renderToStaticMarkup(React.createElement(PhotoCard, { image: images.find((image) => image.isPrimary), title: "Pig", slug: "pig" }));
  assert.match(card, /https:\/\/signed\.test\/selected/);
  const gallery = renderToStaticMarkup(React.createElement(ProductGallery, { slug: "pig", images: images.map((image) => ({ url: image.url, altText: image.altText, width: 800, height: 800 })) }));
  assert.ok(gallery.indexOf("https://signed.test/first") < gallery.indexOf("https://signed.test/selected"));
});

test("featured catalog cards choose the primary image outside gallery order", async () => {
  const { default: FeaturedProducts } = await loadModule("src/components/store/product/FeaturedProducts.tsx");
  const product = { id: "pig", slug: "pig", title: "Pig", price: 10, images: [
    { id: "first", url: "https://signed.test/first", displayOrder: 0, isPrimary: false },
    { id: "cover", url: "https://signed.test/cover", displayOrder: 1, isPrimary: true },
  ] };
  const markup = renderToStaticMarkup(React.createElement(FeaturedProducts, { products: [product] }));
  assert.match(markup, /https:\/\/signed\.test\/cover/);
  assert.doesNotMatch(markup, /https:\/\/signed\.test\/first/);
});
