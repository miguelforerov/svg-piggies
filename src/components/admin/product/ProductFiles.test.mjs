import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { test } from "node:test";
import vm from "node:vm";
import ts from "typescript";

const fileRecord = { id: "file-id", productId: "product-id", fileName: "pig.svg", displayName: "Pig SVG",
  format: "svg", contentType: "image/svg+xml", fileSizeBytes: 11, availabilityStatus: "active",
  displayOrder: 0, createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" };

function component(name) {
  const value = () => {};
  value.displayName = name;
  return value;
}

function deferred() {
  let resolve;
  let reject;
  const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}

function harness({ initialFiles = [fileRecord], getRequest, uploadRequest, deleteRequest } = {}) {
  const state = [];
  let index = 0;
  let effects = [];
  let files = initialFiles;
  let getCalls = 0;
  let uploadCalls = 0;
  let deleteCalls = 0;
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
    "Button", "Badge", "AlertDialog", "AlertDialogAction", "AlertDialogCancel", "AlertDialogContent",
    "AlertDialogDescription", "AlertDialogFooter", "AlertDialogHeader", "AlertDialogTitle",
  ].map((name) => [name, component(name)]));
  const imports = {
    react,
    "react/jsx-runtime": { jsx, jsxs: jsx },
    "lucide-react": { FileUp: component("FileUp"), RefreshCw: component("RefreshCw"), Trash2: component("Trash2") },
    "@/components/ui/button": { Button: ui.Button },
    "@/components/ui/badge": { Badge: ui.Badge },
    "@/components/ui/alert-dialog": ui,
    "@/lib/product-files": {
      getProductFiles: async (...args) => { getCalls++; return getRequest ? getRequest(...args) : files; },
      uploadProductFile: async (...args) => {
        uploadCalls++;
        const confirmed = await (uploadRequest ? uploadRequest(...args) : fileRecord);
        files = [confirmed, ...files.filter((item) => item.id !== confirmed.id)];
        return confirmed;
      },
      deleteProductFile: async (...args) => {
        deleteCalls++;
        await deleteRequest?.(...args);
        files = files.filter((item) => item.id !== args[1]);
      },
      validateProductFile: (file) => file.name.endsWith(".zip") ? "ZIP is not supported." : null,
    },
  };
  const source = readFileSync(new URL("./ProductFiles.tsx", import.meta.url), "utf8");
  const compiled = ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX, target: ts.ScriptTarget.ES2022 },
  }).outputText;
  const module = { exports: {} };
  vm.runInNewContext(compiled, { module, exports: module.exports, require: (name) => {
    if (!(name in imports)) throw new Error(`Unexpected import ${name}`);
    return imports[name];
  }, Set, Error, console });
  const ProductFiles = module.exports.ProductFiles;
  function render() {
    index = 0;
    effects = [];
    const tree = ProductFiles({ productId: "product-id" });
    effects.forEach((effect) => effect());
    return tree;
  }
  function all(tree, predicate) {
    if (Array.isArray(tree)) return tree.flatMap((child) => all(child, predicate));
    if (!tree || typeof tree !== "object") return [];
    const found = predicate(tree) ? [tree] : [];
    return found.concat(all(tree.props?.children, predicate));
  }
  const byName = (tree, name) => all(tree, (node) => node.type?.displayName === name);
  const byLabel = (tree, label) => all(tree, (node) => node.props?.["aria-label"] === label);
  const flush = async () => { await new Promise((resolve) => setImmediate(resolve)); };
  return { render, all, byName, byLabel, flush, calls: () => ({ getCalls, uploadCalls, deleteCalls }), setFiles: (next) => { files = next; } };
}

const selected = (name = "pig.svg") => new File(["file bytes"], name, { type: "image/svg+xml" });

test("shows loading, an empty state, and a safe list error", async () => {
  const loading = deferred();
  const app = harness({ initialFiles: [], getRequest: () => loading.promise });
  assert.match(JSON.stringify(app.render()), /Loading files/);
  loading.resolve([]);
  await app.flush();
  assert.match(JSON.stringify(app.render()), /No files yet/);

  const failed = harness({ getRequest: async () => { throw new Error("Could not load files."); } });
  failed.render();
  await failed.flush();
  const tree = failed.render();
  assert.ok(failed.all(tree, (node) => node.props?.role === "alert").length);
  assert.ok(failed.byLabel(tree, "Refresh product files").length);
});

test("lists only returned metadata and offers a file-specific delete action", async () => {
  const app = harness();
  app.render();
  await app.flush();
  const tree = app.render();
  assert.ok(app.byLabel(tree, "Delete Pig SVG").length);
  assert.match(JSON.stringify(tree), /pig\.svg.*SVG.*active/);
  assert.ok(!JSON.stringify(tree).includes("fileKey"));
  assert.ok(!JSON.stringify(tree).includes("uploadUrl"));
});

test("drag and selection queue valid files, reject ZIP, and avoid duplicates", async () => {
  const app = harness({ initialFiles: [] });
  app.render(); await app.flush();
  const valid = selected();
  const invalid = selected("pig.zip");
  const drop = app.all(app.render(), (node) => typeof node.props?.onDrop === "function")[0];
  drop.props.onDrop({ preventDefault() {}, dataTransfer: { files: [valid, invalid] } });
  let tree = app.render();
  assert.ok(app.all(tree, (node) => node.props?.role === "alert").some((node) => JSON.stringify(node).includes("ZIP")));
  assert.equal(app.byName(tree, "Button").filter((node) => node.props.children === "Retry").length, 0);
  const input = app.byLabel(tree, "Select product files")[0];
  assert.equal(input.props.accept.includes(".zip"), false);
  input.props.onChange({ target: { files: [valid], value: "fake" } });
  tree = app.render();
  assert.equal(app.byLabel(tree, "Remove pig.svg from selection").length, 1);
  assert.equal(app.byLabel(tree, "Remove pig.zip from selection").length, 1);
});

test("uploads sequentially, blocks duplicate clicks, and updates the list after success", async () => {
  const first = deferred();
  const app = harness({ initialFiles: [], uploadRequest: (productId, file, onStage) => {
    onStage("uploading");
    return file.name === "pig.svg" ? first.promise : { ...fileRecord, id: "second-id", fileName: file.name, displayName: file.name };
  } });
  app.render(); await app.flush();
  app.byLabel(app.render(), "Select product files")[0].props.onChange({ target: { files: [selected(), selected("second.svg")], value: "fake" } });
  let tree = app.render();
  const upload = app.byName(tree, "Button").find((node) => node.props.children === "Upload");
  upload.props.onClick();
  upload.props.onClick();
  assert.equal(app.calls().uploadCalls, 1);
  tree = app.render();
  assert.equal(app.byName(tree, "Button").find((node) => node.props.children === "Upload").props.disabled, true);
  first.resolve(fileRecord);
  await app.flush();
  assert.equal(app.calls().uploadCalls, 2);
  await app.flush();
  tree = app.render();
  assert.ok(app.byLabel(tree, "Delete Pig SVG").length);
  assert.ok(app.byLabel(tree, "Delete second.svg").length);
  assert.ok(app.all(tree, (node) => node.props?.role === "status").some((node) => JSON.stringify(node).includes("Completed")));
});

test("failed upload remains retryable and retry starts a new upload flow", async () => {
  let fail = true;
  const app = harness({ initialFiles: [], uploadRequest: async () => {
    if (fail) throw new Error("The upload authorization expired. Retry to request a new one.");
    return fileRecord;
  } });
  app.render(); await app.flush();
  app.byLabel(app.render(), "Select product files")[0].props.onChange({ target: { files: [selected()], value: "fake" } });
  app.byName(app.render(), "Button").find((node) => node.props.children === "Upload").props.onClick();
  await app.flush();
  await app.flush();
  let tree = app.render();
  assert.ok(app.all(tree, (node) => node.props?.role === "alert").some((node) => JSON.stringify(node).includes("expired")), JSON.stringify(tree));
  fail = false;
  app.byName(tree, "Button").find((node) => node.props.children === "Retry").props.onClick();
  await app.flush();
  assert.equal(app.calls().uploadCalls, 2);
  tree = app.render();
  assert.ok(app.byLabel(tree, "Delete Pig SVG").length);
});

test("delete confirmation waits for success, blocks duplicates, and removes only afterward", async () => {
  const deletion = deferred();
  const app = harness({ deleteRequest: () => deletion.promise });
  app.render(); await app.flush();
  app.byLabel(app.render(), "Delete Pig SVG")[0].props.onClick();
  let tree = app.render();
  assert.equal(app.byName(tree, "AlertDialog")[0].props.open, true);
  app.byName(tree, "AlertDialog")[0].props.onOpenChange(false);
  assert.equal(app.calls().deleteCalls, 0);
  app.byLabel(app.render(), "Delete Pig SVG")[0].props.onClick();
  tree = app.render();
  const confirm = app.byName(tree, "AlertDialogAction")[0];
  confirm.props.onClick(); confirm.props.onClick();
  assert.equal(app.calls().deleteCalls, 1);
  assert.equal(app.byLabel(app.render(), "Delete Pig SVG").length, 1);
  deletion.resolve();
  await app.flush();
  await app.flush();
  assert.equal(app.byLabel(app.render(), "Delete Pig SVG").length, 0);
});

test("failed deletion retains the file and reports the error in the dialog", async () => {
  const app = harness({ deleteRequest: async () => { throw new Error("File storage is temporarily unavailable."); } });
  app.render(); await app.flush();
  app.byLabel(app.render(), "Delete Pig SVG")[0].props.onClick();
  app.byName(app.render(), "AlertDialogAction")[0].props.onClick();
  await app.flush();
  await app.flush();
  const tree = app.render();
  assert.ok(app.byLabel(tree, "Delete Pig SVG").length);
  assert.equal(app.byName(tree, "AlertDialog")[0].props.open, true);
  assert.ok(app.all(tree, (node) => node.props?.role === "alert").some((node) => JSON.stringify(node).includes("unavailable")), JSON.stringify(tree));
});
