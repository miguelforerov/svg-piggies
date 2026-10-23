import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import ts from "typescript";

const root = process.cwd();

const sourcePath = path.join(root, "src/types/store.ts");
const outputPath = path.join(
  root,
  "src/styles/generated-design-tokens.css",
);

const sourceText = await readFile(sourcePath, "utf8");

const sourceFile = ts.createSourceFile(
  sourcePath,
  sourceText,
  ts.ScriptTarget.Latest,
  true,
  ts.ScriptKind.TS,
);

function unwrapExpression(node) {
  while (
    ts.isAsExpression(node) ||
    ts.isSatisfiesExpression(node) ||
    ts.isParenthesizedExpression(node) ||
    ts.isTypeAssertionExpression(node)
  ) {
    node = node.expression;
  }

  return node;
}

function getPropertyName(name) {
  if (
    ts.isIdentifier(name) ||
    ts.isStringLiteral(name) ||
    ts.isNumericLiteral(name)
  ) {
    return name.text;
  }

  throw new Error("Nombre de propiedad no compatible en design.");
}

function readObject(node) {
  node = unwrapExpression(node);

  if (!ts.isObjectLiteralExpression(node)) {
    throw new Error("La configuración design debe ser un objeto.");
  }

  const result = {};

  for (const property of node.properties) {
    if (!ts.isPropertyAssignment(property)) {
      throw new Error(
        "design solo admite propiedades con valores explícitos.",
      );
    }

    const key = getPropertyName(property.name);
    const value = unwrapExpression(property.initializer);

    if (
      ts.isStringLiteral(value) ||
      ts.isNoSubstitutionTemplateLiteral(value)
    ) {
      result[key] = value.text;
    } else if (ts.isObjectLiteralExpression(value)) {
      result[key] = readObject(value);
    } else {
      throw new Error(
        `Valor no compatible en design.${key}. Usa strings o un objeto.`,
      );
    }
  }

  return result;
}

const declaration = sourceFile.statements
  .filter(ts.isVariableStatement)
  .flatMap((statement) => [
    ...statement.declarationList.declarations,
  ])
  .find(
    (item) =>
      ts.isIdentifier(item.name) && item.name.text === "design",
  );

if (!declaration?.initializer) {
  throw new Error("No se encontró export const design en store.ts.");
}

const design = readObject(declaration.initializer);
const tokens = new Map();

function addToken(name, value) {
  if (tokens.has(name) && tokens.get(name) !== value) {
    throw new Error(
      `El token ${name} tiene dos valores diferentes en design.`,
    );
  }

  tokens.set(name, value);
}

for (const [section, values] of Object.entries(design)) {
  if (!["radius", "spacing", "gap"].includes(section)) {
    continue;
  }

  if (typeof values !== "object" || values === null) {
    throw new Error(`design.${section} debe ser un objeto.`);
  }

  for (const [name, value] of Object.entries(values)) {
    if (typeof value !== "string" || value.trim() === "") {
      throw new Error(
        `design.${section}.${name} debe ser un string no vacío.`,
      );
    }

    if (section === "radius") {
      addToken(`--radius-${name}`, value);
    } else {
      // Tailwind comparte --spacing-* para padding, margin y gap.
      addToken(`--spacing-${name}`, value);
    }
  }
}

const css = [
  "/* Generated automatically from src/types/store.ts. */",
  "@theme {",
  ...[...tokens.entries()].map(
    ([name, value]) => `  ${name}: ${value};`,
  ),
  "}",
  "",
].join("\n");

await mkdir(path.dirname(outputPath), { recursive: true });
await writeFile(outputPath, css, "utf8");

console.log("Generated Design Tokens:", outputPath);