import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { test } from "node:test";
import vm from "node:vm";
import ts from "typescript";

function component(name) {
  const value = () => {};
  value.displayName = name;
  return value;
}

function harness() {
  const state = [];
  let index = 0;
  let effects = [];
  let saved = ["a", "b"];
  let writes = 0;
  let resolveWrite;
  let pendingWrite = false;
  const products = [
    { id: "source", title: "Source Pig", slug: "source-pig" },
    { id: "a", title: "Apple Pig", slug: "apple-pig" },
    { id: "b", title: "Blue Pig", slug: "blue-pig" },
    { id: "c", title: "Cloud Pig", slug: "cloud-pig" },
  ];
  const react = {
    useState(initial) {
      const slot = index++;
      if (!(slot in state)) state[slot] = initial;
      return [state[slot], (value) => { state[slot] = typeof value === "function" ? value(state[slot]) : value; }];
    },
    useRef(initial) {
      const slot = index++;
      if (!(slot in state)) state[slot] = { current: initial };
      return state[slot];
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
  const ui = Object.fromEntries(["Button", "Input"].map((name) => [name, component(name)]));
  const source = readFileSync(new URL("./ProductRelationships.tsx", import.meta.url), "utf8");
  const compiled = ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX, target: ts.ScriptTarget.ES2022 },
  }).outputText;
  const module = { exports: {} };
  const imports = {
    react,
    "react/jsx-runtime": { jsx, jsxs: jsx },
    "lucide-react": Object.fromEntries(["ArrowDown", "ArrowUp", "GripVertical", "Plus", "RefreshCw", "Trash2"].map((name) => [name, component(name)])),
    "@dnd-kit/react": { DragDropProvider: component("DragDropProvider") },
    "@dnd-kit/react/sortable": { useSortable: () => ({ ref: () => {}, handleRef: () => {}, isDragSource: false, isDropTarget: false }), isSortable: () => true },
    "@/components/ui/button": { Button: ui.Button },
    "@/components/ui/input": { Input: ui.Input },
    "@/lib/product-images": { getProductImages: async () => [] },
    "@/lib/product-relationships": {
      getAdminProducts: async () => products,
      getProductRelationships: async () => saved.map((id, displayOrder) => ({ id, relatedProductId: id, displayOrder })),
      orderedRelatedProductIds: (relationships) => [...relationships].sort((a, b) => a.displayOrder - b.displayOrder).map((item) => item.relatedProductId),
      replaceProductRelationships: async (_id, ids) => {
        writes++;
        if (pendingWrite) await new Promise((resolve) => { resolveWrite = resolve; });
        saved = [...ids];
      },
      searchAvailableProducts: (items, query, productId, selectedIds) => items.filter((item) => query && item.id !== productId && !selectedIds.includes(item.id) && (item.title.toLowerCase().includes(query.toLowerCase()) || item.slug.includes(query.toLowerCase()))),
    },
  };
  const listeners = new Map();
  vm.runInNewContext(compiled, {
    module, exports: module.exports,
    require: (name) => {
      if (!(name in imports)) throw new Error(`Unexpected import ${name}`);
      return imports[name];
    },
    window: { addEventListener: (name, callback) => listeners.set(name, callback), removeEventListener: (name) => listeners.delete(name) },
    Map, Set, console,
  });
  const ProductRelationships = module.exports.ProductRelationships;
  function render() {
    index = 0;
    effects = [];
    const tree = ProductRelationships({ productId: "source" });
    effects.forEach((effect) => effect());
    return tree;
  }
  function all(tree, predicate) {
    if (Array.isArray(tree)) return tree.flatMap((child) => all(child, predicate));
    if (!tree || typeof tree !== "object") return [];
    const found = predicate(tree) ? [tree] : [];
    const children = tree.type?.name === "SortableRelatedProduct" ? tree.type(tree.props) : tree.props?.children;
    return found.concat(all(children, predicate));
  }
  const byLabel = (tree, label) => all(tree, (node) => node.props?.["aria-label"] === label);
  const byName = (tree, name) => all(tree, (node) => node.type?.displayName === name);
  const flush = async () => { await new Promise((resolve) => setImmediate(resolve)); };
  return { render, byLabel, byName, flush, writes: () => writes, saved: () => saved, listeners, deferWrite: () => { pendingWrite = true; }, completeWrite: () => { pendingWrite = false; resolveWrite(); } };
}

test("shows existing relationships, edits the draft, and confirms the saved order", async () => {
  const app = harness();
  app.render();
  await app.flush();
  let tree = app.render();
  assert.ok(app.byLabel(tree, "Remove Apple Pig").length);
  app.byLabel(tree, "Move Blue Pig up")[0].props.onClick();
  tree = app.render();
  app.byLabel(tree, "Remove Apple Pig")[0].props.onClick();
  tree = app.render();
  app.byName(tree, "Input")[0].props.onChange({ target: { value: "cloud" } });
  tree = app.render();
  app.byLabel(tree, "Add Cloud Pig")[0].props.onClick();
  tree = app.render();
  assert.equal(app.byLabel(tree, "Add Cloud Pig").length, 0, "selected product must not remain available");
  assert.ok(app.listeners.has("beforeunload"), "unsaved changes must warn on navigation");
  app.deferWrite();
  const save = app.byName(tree, "Button").find((button) => button.props.children === "Save related products");
  save.props.onClick();
  tree = app.render();
  assert.ok(app.byName(tree, "Button").some((button) => button.props.children === "Saving..."));
  save.props.onClick();
  assert.equal(app.writes(), 1, "saving must reject duplicate requests");
  app.completeWrite();
  await app.flush();
  tree = app.render();
  assert.deepEqual(app.saved(), ["b", "c"]);
  assert.equal(app.byName(tree, "Button").some((button) => button.props.children === "Save related products"), false);
  assert.equal(app.listeners.has("beforeunload"), false);
});
