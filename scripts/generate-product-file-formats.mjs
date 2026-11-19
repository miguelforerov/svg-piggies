import { readFile, writeFile } from "node:fs/promises";
const source = await readFile(new URL("../config/product-file-formats.json", import.meta.url), "utf8");
const config = JSON.parse(source);
if (config === null || typeof config !== "object" || Array.isArray(config) ||
    Object.keys(config).some((key) => key !== "formats" && key !== "_comment") ||
    (Object.hasOwn(config, "_comment") && typeof config._comment !== "string") ||
    !Array.isArray(config.formats) || config.formats.length === 0 ||
    config.formats.some(({ format, extension, contentTypes, ...extra }) =>
      Object.keys(extra).length > 0 || !/^[a-z][a-z0-9]*$/.test(format) || extension !== `.${format}` ||
      !Array.isArray(contentTypes) || contentTypes.length === 0 ||
      contentTypes.some((type) => typeof type !== "string" || !/^[a-z0-9.+-]+\/[a-z0-9.+-]+$/.test(type)) ||
      new Set(contentTypes).size !== contentTypes.length) ||
    new Set(config.formats.map(({ format }) => format)).size !== config.formats.length) {
  throw new Error("Invalid product file format configuration");
}

const goSource = `// Code generated from config/product-file-formats.json; DO NOT EDIT.\npackage productfiles\n\nconst embeddedProductFileFormatsJSON = ${JSON.stringify(JSON.stringify({ formats: config.formats }))}\n`;
const tsSource = `// Code generated from config/product-file-formats.json; DO NOT EDIT.\nexport type ProductFileFormat = ${config.formats.map(({ format }) => JSON.stringify(format)).join(" | ")};\n`;
await writeFile(new URL("../backend/internal/productfiles/formats_generated.go", import.meta.url), goSource);
await writeFile(new URL("../src/lib/product-file-formats.generated.ts", import.meta.url), tsSource);
