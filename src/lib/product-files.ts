import formatConfig from "../../config/product-file-formats.json" with { type: "json" };
import type { ProductFileFormat } from "./product-file-formats.generated";

export type { ProductFileFormat } from "./product-file-formats.generated";
export type ProductFileAvailability = "active" | "retired" | "blocked";

export interface AdminProductFile {
  id: string;
  productId: string;
  fileName: string;
  displayName: string;
  format: ProductFileFormat;
  contentType: string;
  fileSizeBytes: number;
  availabilityStatus: ProductFileAvailability;
  displayOrder: number;
  createdAt: string;
  updatedAt: string;
}

interface UploadAuthorization {
  uploadToken: string;
  uploadUrl: string;
  headers: Record<string, string[]>;
  expiresAt: string;
}

const maxFileBytes = 50_000_000;
const definitions = formatConfig.formats;
const byFormat = new Map(definitions.map((definition) => [definition.format, definition]));
export const acceptedProductFileExtensions = definitions.map((definition) => definition.extension).join(",");
const formatLabels = definitions.map((definition) => definition.format.toUpperCase());
export const supportedProductFileLabel = formatLabels.length === 1 ? formatLabels[0]
  : `${formatLabels.slice(0, -1).join(", ")}, or ${formatLabels.at(-1)}`;

export class ProductFileApiError extends Error {
  readonly status: number;

  constructor(status: number) {
    super(fileErrorMessage(status));
    this.status = status;
  }
}

function fileErrorMessage(status: number): string {
  switch (status) {
    case 400: return "The file details are invalid. Check the file and try again.";
    case 401: return "Your admin session has expired. Sign in again.";
    case 403: return "You do not have permission to manage files.";
    case 404: return "The product or file was not found.";
    case 409: return "The file could not be verified or conflicts with an existing upload. Try again.";
    case 410: return "The upload authorization expired. Retry to request a new one.";
    case 502:
    case 503: return "File storage is temporarily unavailable. Try again.";
    default: return "The file request failed. Try again.";
  }
}

const pathFor = (productId: string) => `/api/admin/products/${encodeURIComponent(productId)}/files`;

async function jsonRequest<T>(path: string, expectedStatus: number, init?: RequestInit): Promise<T> {
  let response: Response;
  try {
    response = await fetch(path, {
      ...init,
      headers: { Accept: "application/json", ...(init?.body ? { "Content-Type": "application/json" } : {}) },
      cache: "no-store",
    });
  } catch {
    throw new ProductFileApiError(0);
  }
  if (response.status !== expectedStatus) throw new ProductFileApiError(response.status);
  try {
    return await response.json() as T;
  } catch {
    throw new ProductFileApiError(0);
  }
}

function fileFormat(fileName: string): ProductFileFormat | null {
  const name = fileName.toLowerCase();
  return (definitions.find((definition) => name.endsWith(definition.extension))?.format as ProductFileFormat | undefined) ?? null;
}

function fileContentType(file: File, format: ProductFileFormat): string {
  // Browsers often leave DXF (and occasionally other formats) without a MIME type.
  // An explicit, incompatible MIME type is never replaced.
  return file.type || byFormat.get(format)!.contentTypes[0];
}

export function validateProductFile(file: File): string | null {
  const name = file.name;
  if (!name || new TextEncoder().encode(name).length > 255 || name !== name.trim() ||
      /[/\\\p{Cc}]/u.test(name)) {
    return "Filename must be 1–255 bytes with no surrounding spaces, path separators, or control characters.";
  }
  const format = fileFormat(name);
  if (!format) return `Use an ${supportedProductFileLabel} file.`;
  if (!byFormat.get(format)!.contentTypes.includes(fileContentType(file, format))) {
    return "The file extension and MIME type do not match an allowed format.";
  }
  if (file.size < 1 || file.size > maxFileBytes) {
    return "File size must be between 1 and 50,000,000 bytes (50 MB).";
  }
  return null;
}

export async function getProductFiles(productId: string): Promise<AdminProductFile[]> {
  const files = await jsonRequest<AdminProductFile[]>(pathFor(productId), 200);
  if (!Array.isArray(files)) throw new ProductFileApiError(0);
  return files;
}

export async function deleteProductFile(productId: string, fileId: string): Promise<void> {
  let response: Response;
  try {
    response = await fetch(`${pathFor(productId)}/${encodeURIComponent(fileId)}`, {
      method: "DELETE",
      headers: { Accept: "application/json" },
      cache: "no-store",
    });
  } catch {
    throw new ProductFileApiError(0);
  }
  if (response.status !== 204) throw new ProductFileApiError(response.status);
}

export async function uploadProductFile(
  productId: string,
  file: File,
  onStage: (stage: "authorizing" | "uploading" | "confirming") => void,
): Promise<AdminProductFile> {
  const validationError = validateProductFile(file);
  if (validationError) throw new Error(validationError);
  const format = fileFormat(file.name)!;
  const contentType = fileContentType(file, format);

  onStage("authorizing");
  const authorization = await jsonRequest<UploadAuthorization>(`${pathFor(productId)}/upload-authorization`, 201, {
    method: "POST",
    body: JSON.stringify({ fileName: file.name, format, contentType, fileSizeBytes: file.size }),
  });
  const expiresAt = Date.parse(authorization.expiresAt);
  if (!Number.isFinite(expiresAt)) throw new ProductFileApiError(0);
  if (expiresAt <= Date.now()) throw new ProductFileApiError(410);

  let headers: Headers;
  try {
    headers = new Headers();
    for (const [name, values] of Object.entries(authorization.headers)) {
      if (!Array.isArray(values) || values.length === 0 || values.some((value) => typeof value !== "string")) {
        throw new Error("invalid signed headers");
      }
      for (const value of values) headers.append(name, value);
    }
    const uploadURL = new URL(authorization.uploadUrl);
    if (uploadURL.protocol !== "https:" || uploadURL.username || uploadURL.password ||
        headers.get("Content-Type") !== contentType || headers.get("If-None-Match") !== "*") {
      throw new Error("invalid upload authorization");
    }
  } catch {
    throw new ProductFileApiError(0);
  }

  onStage("uploading");
  let upload: Response;
  try {
    upload = await fetch(authorization.uploadUrl, {
      method: "PUT", headers, body: file, credentials: "omit", redirect: "error",
    });
  } catch {
    throw new Error("Direct file upload failed. Check the files bucket CORS policy and your connection, then retry.");
  }
  if (!upload.ok) throw new Error("Direct file upload failed. Retry to request a new authorization.");
  if (expiresAt <= Date.now()) throw new ProductFileApiError(410);

  onStage("confirming");
  return jsonRequest<AdminProductFile>(pathFor(productId), 201, {
    method: "POST",
    body: JSON.stringify({ uploadToken: authorization.uploadToken }),
  });
}
