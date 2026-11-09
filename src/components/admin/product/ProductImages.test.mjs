import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { test } from "node:test";
import vm from "node:vm";
import ts from "typescript";

const image = {
  id: "image-id", productId: "product-id", originalFilename: "pig.webp",
  imageUrl: "https://example.test/pig.webp", displayOrder: 0, isPrimary: true,
  fileSizeBytes: 100, altText: "Pig",
};
const secondaryImage = { ...image, id: "secondary-id", originalFilename: "second.webp", displayOrder: 1, isPrimary: false };

function component(name) {
  const value = () => {};
  value.displayName = name;
  return value;
}

function harness(deleteRequest, primaryRequest = async () => {}, altRequest = async () => {}, orderRequest = async () => {}) {
  const state = [];
  let index = 0;
  let effects = [];
  let list = [image];
  let deleteCalls = 0;
  let primaryCalls = 0;
  let altCalls = 0;
  let orderCalls = 0;
  let lastOrder;
  let lastAltArgs;
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
    useEffect(effect) {
      const slot = index++;
      if (!(slot in state)) { state[slot] = true; effects.push(effect); }
    },
  };
  const jsx = (type, props) => ({ type, props: props ?? {} });
  const ui = Object.fromEntries([
    "Button", "Input", "Textarea", "Badge", "Dialog", "DialogContent", "DialogDescription", "DialogHeader", "DialogTitle", "AlertDialog", "AlertDialogAction", "AlertDialogCancel",
    "AlertDialogContent", "AlertDialogDescription", "AlertDialogFooter",
    "AlertDialogHeader", "AlertDialogTitle",
  ].map((name) => [name, component(name)]));
  const source = readFileSync(new URL("./ProductImages.tsx", import.meta.url), "utf8");
  const compiled = ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX, target: ts.ScriptTarget.ES2022 },
  }).outputText;
  const module = { exports: {} };
  const imports = {
    react,
    "react/jsx-runtime": { jsx, jsxs: jsx },
    "lucide-react": { ImagePlus: component("ImagePlus"), RefreshCw: component("RefreshCw"), Trash2: component("Trash2"), GripVertical: component("GripVertical") },
    "@dnd-kit/react": { DragDropProvider: component("DragDropProvider") },
    "@dnd-kit/react/sortable": { useSortable: () => ({ ref: () => {}, handleRef: () => {}, isDragSource: false, isDropTarget: false }), isSortable: () => true },
    "@/components/ui/button": { Button: ui.Button },
    "@/components/ui/input": { Input: ui.Input },
    "@/components/ui/textarea": { Textarea: ui.Textarea },
    "@/components/ui/dialog": ui,
    "@/components/ui/badge": { Badge: ui.Badge },
    "@/components/ui/alert-dialog": ui,
    "@/lib/product-images": {
      ProductImageApiError: class ProductImageApiError extends Error { constructor(status) { super("Image request failed"); this.status = status; } },
      getProductImages: async () => list,
      deleteProductImage: async (...args) => { deleteCalls++; return deleteRequest(...args); },
      setPrimaryProductImage: async (...args) => { primaryCalls++; return primaryRequest(...args); },
      updateProductImageAltText: async (...args) => { altCalls++; lastAltArgs = args; return altRequest(...args); },
      reorderProductImages: async (...args) => { orderCalls++; lastOrder = args; return orderRequest(...args); },
      uploadProductImage: async () => {},
      validateImage: () => null,
    },
  };
  vm.runInNewContext(compiled, { module, exports: module.exports, require: (name) => {
    if (!(name in imports)) throw new Error(`Unexpected import ${name}`);
    return imports[name];
  }, URL, Set, console });
  const ProductImages = module.exports.ProductImages;
  function render() {
    index = 0;
    effects = [];
    const tree = ProductImages({ productId: "product-id" });
    effects.forEach((effect) => effect());
    return tree;
  }
  function all(tree, predicate) {
    if (Array.isArray(tree)) return tree.flatMap((child) => all(child, predicate));
    if (!tree || typeof tree !== "object") return [];
    const found = predicate(tree) ? [tree] : [];
    const children = tree.type?.name === "SortableImage" ? tree.type(tree.props) : tree.props?.children;
    return found.concat(all(children, predicate));
  }
  const byName = (tree, name) => all(tree, (node) => node.type?.displayName === name);
  const byLabel = (tree, label) => all(tree, (node) => node.props?.["aria-label"] === label);
  const flush = async () => { await new Promise((resolve) => setImmediate(resolve)); };
  return { render, byName, byLabel, all, flush, calls: () => deleteCalls, primaryCalls: () => primaryCalls, altCalls: () => altCalls, lastAltArgs: () => lastAltArgs, orderCalls: () => orderCalls, lastOrder: () => lastOrder, setList: (value) => { list = value; } };
}

test("shows an image-specific delete action and confirmation; cancel sends no request", async () => {
  const app = harness(async () => {});
  app.render();
  await app.flush();
  let tree = app.render();
  const button = app.byLabel(tree, "Delete pig.webp")[0];
  assert.ok(button, "delete action should appear beside the image");
  button.props.onClick();
  tree = app.render();
  assert.equal(app.byName(tree, "AlertDialog")[0].props.open, true);
  assert.match(JSON.stringify(tree), /pig\.webp/);
  assert.ok(app.byName(tree, "AlertDialogCancel").length);
  app.byName(tree, "AlertDialog")[0].props.onOpenChange(false);
  tree = app.render();
  assert.equal(app.byName(tree, "AlertDialog")[0].props.open, false);
  assert.equal(app.calls(), 0);
});

test("successful confirmation waits for the response, blocks duplicates, and updates the gallery", async () => {
  let resolveDelete;
  const app = harness(() => new Promise((resolve) => { resolveDelete = resolve; }));
  app.render();
  await app.flush();
  app.byLabel(app.render(), "Delete pig.webp")[0].props.onClick();
  let tree = app.render();
  const confirm = app.byName(tree, "AlertDialogAction")[0];
  confirm.props.onClick();
  confirm.props.onClick();
  assert.equal(app.calls(), 1);
  tree = app.render();
  assert.equal(app.byName(tree, "AlertDialogAction")[0].props.disabled, true);
  assert.equal(app.byLabel(tree, "Delete pig.webp").length, 1, "image stays visible while request is pending");
  resolveDelete();
  await app.flush();
  tree = app.render();
  assert.equal(app.byLabel(tree, "Delete pig.webp").length, 0);
  assert.ok(app.all(tree, (node) => node.props?.role === "status").some((node) => JSON.stringify(node).includes("was deleted")));
});

test("failed confirmation keeps the image and shows an error", async () => {
  const app = harness(async () => { throw new Error("Storage is unavailable"); });
  app.render();
  await app.flush();
  app.byLabel(app.render(), "Delete pig.webp")[0].props.onClick();
  app.byName(app.render(), "AlertDialogAction")[0].props.onClick();
  await app.flush();
  const tree = app.render();
  assert.equal(app.byLabel(tree, "Delete pig.webp").length, 1);
  assert.ok(app.all(tree, (node) => node.props?.role === "alert").some((node) => JSON.stringify(node).includes("Could not delete the image")));
});

test("identifies the primary image and offers selection only for secondary images", async () => {
  const app = harness(async () => {});
  app.setList([image, secondaryImage]);
  app.render();
  await app.flush();
  const tree = app.render();
  assert.equal(app.byName(tree, "Badge").filter((node) => node.props.children === "Primary").length, 1);
  assert.equal(app.byLabel(tree, "Set pig.webp as primary").length, 0);
  assert.equal(app.byLabel(tree, "Set second.webp as primary").length, 1);
});

test("selection waits for 204, blocks duplicates, and changes the primary badge", async () => {
  let resolvePrimary;
  const app = harness(async () => {}, () => new Promise((resolve) => { resolvePrimary = resolve; }));
  app.setList([image, secondaryImage]);
  app.render();
  await app.flush();
  app.byLabel(app.render(), "Set second.webp as primary")[0].props.onClick();
  let tree = app.render();
  assert.equal(app.primaryCalls(), 1);
  assert.equal(app.byLabel(tree, "Set second.webp as primary")[0].props.disabled, true);
  assert.equal(app.byLabel(tree, "Set pig.webp as primary").length, 0, "previous primary remains until success");
  app.byLabel(tree, "Set second.webp as primary")[0].props.onClick();
  assert.equal(app.primaryCalls(), 1);
  resolvePrimary();
  await app.flush();
  tree = app.render();
  assert.equal(app.byLabel(tree, "Set second.webp as primary").length, 0);
  assert.equal(app.byLabel(tree, "Set pig.webp as primary").length, 1);
  assert.ok(app.all(tree, (node) => node.props?.role === "status").some((node) => JSON.stringify(node).includes("primary image")));
});

test("failed selection preserves the previous primary and reports an error", async () => {
  const app = harness(async () => {}, async () => { throw new Error("database unavailable"); });
  app.setList([image, secondaryImage]);
  app.render();
  await app.flush();
  app.byLabel(app.render(), "Set second.webp as primary")[0].props.onClick();
  await app.flush();
  const tree = app.render();
  assert.equal(app.byLabel(tree, "Set pig.webp as primary").length, 0);
  assert.equal(app.byLabel(tree, "Set second.webp as primary").length, 1);
  assert.ok(app.all(tree, (node) => node.props?.role === "alert").some((node) => JSON.stringify(node).includes("Could not set the primary image")));
});

test("alt editor opens with current text and cancel preserves it without a request", async () => {
  const app = harness(async () => {});
  app.render();
  await app.flush();
  app.byLabel(app.render(), "Edit alt text for pig.webp")[0].props.onClick();
  let tree = app.render();
  assert.equal(app.byName(tree, "Dialog")[0].props.open, true);
  assert.equal(app.byLabel(tree, "Alternative text")[0].props.value, "Pig");
  app.byLabel(tree, "Alternative text")[0].props.onChange({ target: { value: "changed" } });
  tree = app.render();
  app.byName(tree, "Button").find((node) => node.props.children === "Cancel").props.onClick();
  tree = app.render();
  assert.equal(app.byName(tree, "Dialog")[0].props.open, false);
  assert.equal(app.altCalls(), 0);
  assert.ok(app.all(tree, (node) => node.type === "img" && node.props?.alt === "Pig").length);
});

test("alt editor displays a null value as an empty draft", async () => {
  const app = harness(async () => {});
  app.setList([{ ...image, altText: null }]);
  app.render();
  await app.flush();
  app.byLabel(app.render(), "Edit alt text for pig.webp")[0].props.onClick();
  assert.equal(app.byLabel(app.render(), "Alternative text")[0].props.value, "");
});

test("alt editor saves only after success, blocks duplicates, and accepts empty text", async () => {
  let resolveUpdate;
  const app = harness(async () => {}, async () => {}, () => new Promise((resolve) => { resolveUpdate = resolve; }));
  app.render();
  await app.flush();
  app.byLabel(app.render(), "Edit alt text for pig.webp")[0].props.onClick();
  let tree = app.render();
  app.byLabel(tree, "Alternative text")[0].props.onChange({ target: { value: "   " } });
  tree = app.render();
  const save = app.byName(tree, "Button").find((node) => node.props.children === "Save");
  save.props.onClick();
  save.props.onClick();
  assert.equal(app.altCalls(), 1);
  assert.deepEqual(app.lastAltArgs(), ["product-id", "image-id", ""]);
  tree = app.render();
  assert.ok(app.byName(tree, "Button").some((node) => node.props.children === "Saving..." && node.props.disabled));
  assert.ok(app.all(tree, (node) => node.type === "img" && node.props?.alt === "Pig").length, "old alt remains until success");
  resolveUpdate();
  await app.flush();
  tree = app.render();
  assert.equal(app.byName(tree, "Dialog")[0].props.open, false);
  assert.ok(app.all(tree, (node) => node.type === "img" && node.props?.alt === "").length);
  assert.ok(app.all(tree, (node) => node.props?.role === "status").some((node) => JSON.stringify(node).includes("was saved")));
});

test("alt editor keeps old text after failure and rejects more than 500 Unicode characters", async () => {
  const app = harness(async () => {}, async () => {}, async () => { throw new Error("database unavailable"); });
  app.render();
  await app.flush();
  app.byLabel(app.render(), "Edit alt text for pig.webp")[0].props.onClick();
  let tree = app.render();
  app.byLabel(tree, "Alternative text")[0].props.onChange({ target: { value: "🐷".repeat(501) } });
  tree = app.render();
  app.byName(tree, "Button").find((node) => node.props.children === "Save").props.onClick();
  assert.equal(app.altCalls(), 0);
  tree = app.render();
  assert.ok(app.all(tree, (node) => node.props?.role === "alert").some((node) => JSON.stringify(node).includes("500 characters")));
  app.byLabel(tree, "Alternative text")[0].props.onChange({ target: { value: "New text" } });
  tree = app.render();
  app.byName(tree, "Button").find((node) => node.props.children === "Save").props.onClick();
  await app.flush();
  tree = app.render();
  assert.equal(app.byName(tree, "Dialog")[0].props.open, true);
  assert.ok(app.all(tree, (node) => node.type === "img" && node.props?.alt === "Pig").length);
  assert.ok(app.all(tree, (node) => node.props?.role === "alert").length);
});

function dragSecondFirst(app) {
  const provider = app.byName(app.render(), "DragDropProvider")[0];
  provider.props.onDragEnd({ canceled: false, operation: { source: { initialIndex: 1, index: 0 } } });
}

test("Grip handle changes provisional order; Cancel restores it without a request", async () => {
  const app = harness(async () => {});
  app.setList([image, secondaryImage]);
  app.render(); await app.flush();
  let tree = app.render();
  assert.equal(app.byLabel(tree, "Move pig.webp").length, 1);
  assert.equal(app.byLabel(tree, "Move second.webp").length, 1);
  assert.equal(typeof app.byLabel(tree, "Move pig.webp")[0].props.ref, "function");
  assert.equal(app.byLabel(tree, "Delete pig.webp")[0].props.ref, undefined);
  assert.equal(app.byName(tree, "DragDropProvider").length, 1);
  dragSecondFirst(app);
  tree = app.render();
  assert.ok(JSON.stringify(tree).includes("Save order"));
  assert.equal(app.byLabel(tree, "Delete pig.webp")[0].props.disabled, true);
  assert.equal(app.byLabel(tree, "Set second.webp as primary")[0].props.disabled, false);
  app.byName(tree, "Button").find((node) => node.props.children === "Cancel").props.onClick();
  tree = app.render();
  assert.equal(app.orderCalls(), 0);
  assert.ok(!JSON.stringify(tree).includes("Save order"));
});

test("Save order waits for success, blocks duplicates, and keeps draft on failure", async () => {
  let resolveOrder;
  const app = harness(async () => {}, async () => {}, async () => {}, () => new Promise((resolve) => { resolveOrder = resolve; }));
  app.setList([image, secondaryImage]);
  app.render(); await app.flush();
  dragSecondFirst(app);
  let tree = app.render();
  const save = app.byName(tree, "Button").find((node) => node.props.children === "Save order");
  save.props.onClick(); save.props.onClick();
  assert.equal(app.orderCalls(), 1);
  assert.deepEqual(Array.from(app.lastOrder()[1]), ["secondary-id", "image-id"]);
  tree = app.render();
  assert.ok(JSON.stringify(tree).includes("Saving..."));
  resolveOrder(); await app.flush();
  tree = app.render();
  assert.ok(!JSON.stringify(tree).includes("Save order"));
  assert.equal(app.byLabel(tree, "Delete pig.webp")[0].props.disabled, false);

  const failed = harness(async () => {}, async () => {}, async () => {}, async () => { throw new Error("Could not save"); });
  failed.setList([image, secondaryImage]); failed.render(); await failed.flush(); dragSecondFirst(failed);
  failed.byName(failed.render(), "Button").find((node) => node.props.children === "Save order").props.onClick();
  await failed.flush();
  assert.ok(JSON.stringify(failed.render()).includes("Save order"));
  assert.ok(failed.all(failed.render(), (node) => node.props?.role === "alert").length);
});

test("changing the primary image preserves an unsaved order", async () => {
  const app = harness(async () => {});
  app.setList([image, secondaryImage]); app.render(); await app.flush(); dragSecondFirst(app);
  app.byLabel(app.render(), "Set second.webp as primary")[0].props.onClick();
  await app.flush();
  const tree = app.render();
  assert.equal(app.byLabel(tree, "Set second.webp as primary").length, 0);
  assert.ok(JSON.stringify(tree).includes("Save order"));
  assert.equal(app.orderCalls(), 0);
});
