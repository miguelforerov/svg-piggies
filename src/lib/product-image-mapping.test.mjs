import assert from "node:assert/strict";
import { test } from "node:test";
import { normalizePublicImages, refreshedImageFor } from "./product-image-mapping.ts";

test("display order is independent of primary and retains alt text and dimensions", () => {
  const images = normalizePublicImages([
    { id: "second", imageUrl: "https://signed/second", displayOrder: 0, isPrimary: false, altText: "Second" },
    { id: "primary", imageUrl: "https://signed/primary", displayOrder: 2, isPrimary: true, altText: "Main pig", widthPx: 800, heightPx: 600 },
  ]);
  assert.equal(images[0].id, "second");
  assert.equal(images[1].id, "primary");
  assert.equal(images[1].altText, "Main pig");
  assert.equal(images[1].widthPx, 800);
  assert.equal(images[1].heightPx, 600);
});

test("empty product leaves image list empty for the placeholder", () => {
  assert.deepEqual(normalizePublicImages([]), []);
});

test("signed URL refresh keeps the same image after primary order changes", () => {
  const refreshed = [{ id: "new-primary" }, { id: "old-primary" }];
  assert.equal(refreshedImageFor(refreshed, "old-primary", 0)?.id, "old-primary");
  assert.equal(refreshedImageFor(refreshed, undefined, 0)?.id, "new-primary");
});
