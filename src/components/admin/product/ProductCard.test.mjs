import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { test } from "node:test";
import vm from "node:vm";
import ts from "typescript";

const component = (name) => Object.assign(() => {}, { displayName: name });
const source = readFileSync(new URL("./ProductCard.tsx", import.meta.url), "utf8");
const compiled = ts.transpileModule(source, {
  compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX, target: ts.ScriptTarget.ES2022 },
}).outputText;

function cardHarness(images) {
  const state = [];
  let index = 0;
  let effects = [];
  let requestedId;
  const react = {
    useState(initial) {
      const slot = index++;
      if (!(slot in state)) state[slot] = initial;
      return [state[slot], (value) => { state[slot] = value; }];
    },
    useEffect(effect, deps) {
      const slot = index++;
      const previous = state[slot];
      if (!previous || deps.some((value, position) => value !== previous.deps[position])) {
        previous?.cleanup?.();
        state[slot] = { deps, cleanup: null };
        effects.push(() => { state[slot].cleanup = effect(); });
      }
    },
  };
  const jsx = (type, props) => ({ type, props: props ?? {} });
  const imports = {
    react,
    "react/jsx-runtime": { jsx, jsxs: jsx },
    "lucide-react": { Package: component("Package") },
    "@/components/admin/product/ProductStatusBadge": { ProductStatusBadge: component("ProductStatusBadge") },
    "@/components/ui/button": { Button: component("Button") },
    "@/components/ui/card": { Card: component("Card"), CardContent: component("CardContent") },
    "@/lib/product-images": { getProductImages: async (id) => { requestedId = id; if (images instanceof Error) throw images; return images; } },
  };
  const module = { exports: {} };
  vm.runInNewContext(compiled, {
    module, exports: module.exports,
    require: (name) => {
      if (!(name in imports)) throw new Error(`Unexpected import: ${name}`);
      return imports[name];
    },
  });
  const product = { id: "product-id", title: "Pig", price: "4.99", status: "active" };
  function render() {
    index = 0;
    effects = [];
    const tree = module.exports.ProductCard({ product });
    effects.forEach((effect) => effect());
    return tree;
  }
  function all(tree, predicate) {
    if (Array.isArray(tree)) return tree.flatMap((child) => all(child, predicate));
    if (!tree || typeof tree !== "object") return [];
    return [...(predicate(tree) ? [tree] : []), ...all(tree.props?.children, predicate)];
  }
  return { render, all, requestedId: () => requestedId };
}

test("admin product card shows the primary image and falls back if it cannot load", async () => {
  const card = cardHarness([
    { imageUrl: "https://example.test/first", isPrimary: false },
    { imageUrl: "https://example.test/primary", isPrimary: true },
  ]);
  card.render();
  await new Promise((resolve) => setImmediate(resolve));
  let tree = card.render();
  assert.equal(card.requestedId(), "product-id");
  const image = card.all(tree, (node) => node.type === "img")[0];
  assert.equal(image.props.src, "https://example.test/primary");
  image.props.onError();
  tree = card.render();
  assert.equal(card.all(tree, (node) => node.type?.displayName === "Package").length, 1);
});

test("admin product card uses the first image when none is primary", async () => {
  const card = cardHarness([{ imageUrl: "https://example.test/first", isPrimary: false }]);
  card.render();
  await new Promise((resolve) => setImmediate(resolve));
  assert.equal(card.all(card.render(), (node) => node.type === "img")[0].props.src, "https://example.test/first");
});

test("admin product card keeps the fallback for empty or failed image requests", async () => {
  for (const images of [[], new Error("unavailable")]) {
    const card = cardHarness(images);
    card.render();
    await new Promise((resolve) => setImmediate(resolve));
    assert.equal(card.all(card.render(), (node) => node.type?.displayName === "Package").length, 1);
  }
});
