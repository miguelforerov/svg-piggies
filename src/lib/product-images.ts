export interface ProductImageRecord {
  id: string;
  productId: string;
  originalFilename: string;
  contentType: string;
  fileSizeBytes: number;
  widthPx?: number | null;
  heightPx?: number | null;
  altText?: string | null;
  displayOrder: number;
  isPrimary: boolean;
  imageUrl: string;
}

interface UploadAuthorization {
  uploadToken: string;
  uploadUrl: string;
  headers: Record<string, string[]>;
  expiresAt: string;
}

export class ProductImageApiError extends Error {
  readonly status: number;
  constructor(status: number) {
    super(imageErrorMessage(status));
    this.status = status;
  }
}

export function imageErrorMessage(status: number): string {
  switch (status) {
    case 400: return "The image details are invalid.";
    case 401: return "Your admin session has expired. Sign in again.";
    case 403: return "You do not have permission to manage images.";
    case 404: return "The product or image was not found.";
    case 409: return "The uploaded image could not be verified. Try again.";
    case 410: return "The upload authorization expired. Try again.";
    case 502: case 503: return "Image storage is temporarily unavailable. Try again.";
    default: return "The image request failed. Try again.";
  }
}

async function jsonRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: { Accept: "application/json", ...(init?.body ? { "Content-Type": "application/json" } : {}) },
    cache: "no-store",
  });
  if (!response.ok) throw new ProductImageApiError(response.status);
  return response.json() as Promise<T>;
}

const pathFor = (productId: string) => `/api/admin/products/${encodeURIComponent(productId)}/images`;

export function validateImage(file: File): string | null {
  if (!["image/jpeg", "image/png", "image/webp"].includes(file.type)) return "Use a JPEG, PNG, or WebP image.";
  if (file.size < 1 || file.size > 10_485_760) return "Image size must be between 1 byte and 10 MiB.";
  if (file.name.length < 1 || file.name.length > 255) return "Filename must be between 1 and 255 characters.";
  return null;
}

export function getProductImages(productId: string): Promise<ProductImageRecord[]> {
  return jsonRequest<ProductImageRecord[]>(pathFor(productId));
}

export async function deleteProductImage(productId: string, imageId: string): Promise<void> {
  const response = await fetch(`${pathFor(productId)}/${encodeURIComponent(imageId)}`, {
    method: "DELETE",
    headers: { Accept: "application/json" },
    cache: "no-store",
  });
  if (response.status !== 204) throw new ProductImageApiError(response.status);
}

export async function setPrimaryProductImage(productId: string, imageId: string): Promise<void> {
  const response = await fetch(`${pathFor(productId)}/${encodeURIComponent(imageId)}/primary`, {
    method: "PUT",
    headers: { Accept: "application/json" },
    cache: "no-store",
  });
  if (response.status !== 204) throw new ProductImageApiError(response.status);
}

export async function reorderProductImages(productId: string, imageIds: string[]): Promise<void> {
  const response = await fetch(`${pathFor(productId)}/order`, {
    method: "PUT",
    headers: { Accept: "application/json", "Content-Type": "application/json" },
    body: JSON.stringify({ imageIds }),
    cache: "no-store",
  });
  if (response.status !== 204) throw new ProductImageApiError(response.status);
}

export async function updateProductImageAltText(productId: string, imageId: string, altText?: string | null): Promise<void> {
  const response = await fetch(`${pathFor(productId)}/${encodeURIComponent(imageId)}`, {
    method: "PATCH",
    headers: { Accept: "application/json", "Content-Type": "application/json" },
    body: JSON.stringify({ altText }),
    cache: "no-store",
  });
  if (response.status !== 204) throw new ProductImageApiError(response.status);
}

export async function uploadProductImage(productId: string, file: File, altText: string, onStage: (stage: "authorizing" | "uploading" | "confirming") => void, placement?: { isPrimary: boolean; displayOrder: number }): Promise<void> {
  const validationError = validateImage(file);
  if (validationError) throw new Error(validationError);

  onStage("authorizing");
  const authorization = await jsonRequest<UploadAuthorization>(`${pathFor(productId)}/uploads`, {
    method: "POST",
    body: JSON.stringify({ originalFilename: file.name, contentType: file.type, fileSizeBytes: file.size }),
  });
  if (Date.parse(authorization.expiresAt) <= Date.now()) throw new ProductImageApiError(410);

  onStage("uploading");
  const headers = new Headers();
  for (const [name, values] of Object.entries(authorization.headers)) {
    for (const value of values) headers.append(name, value);
  }
  let upload: Response;
  try {
    upload = await fetch(authorization.uploadUrl, { method: "PUT", headers, body: file });
  } catch {
    throw new Error("Direct image upload failed. Check the R2 bucket CORS policy and your connection, then try again.");
  }
  if (!upload.ok) throw new Error("Direct image upload failed. Try again.");
  // R2 may return an ETag. Confirmation verifies the object in storage itself.
  upload.headers.get("ETag");
  if (Date.parse(authorization.expiresAt) <= Date.now()) throw new ProductImageApiError(410);

  onStage("confirming");
  await jsonRequest(`${pathFor(productId)}`, {
    method: "POST",
    body: JSON.stringify({ uploadToken: authorization.uploadToken, ...(altText.trim() ? { altText: altText.trim() } : {}), ...(placement ?? {}) }),
  });
}
