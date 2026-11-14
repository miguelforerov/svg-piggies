import type { Product } from "@/components/admin/product/ProductFormContainer";

export interface ProductRelationship {
  id: string;
  productId: string;
  relatedProductId: string;
  displayOrder: number;
}

export class ProductRelationshipApiError extends Error {
  constructor(readonly status: number, message: string) {
    super(message);
    this.name = "ProductRelationshipApiError";
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: {
      Accept: "application/json",
      ...(init?.body ? { "Content-Type": "application/json" } : {}),
    },
    cache: "no-store",
  });
  if (!response.ok) {
    const messages: Record<number, string> = {
      400: "The related product selection is invalid. Refresh and try again.",
      401: "Your admin session has expired. Sign in again.",
      403: "You do not have permission to manage related products.",
      404: "A selected product no longer exists. Refresh and try again.",
      409: "A related product was selected twice. Refresh and try again.",
    };
    throw new ProductRelationshipApiError(
      response.status,
      messages[response.status] ?? "Could not manage related products. Try again.",
    );
  }
  return response.json() as Promise<T>;
}

const pathFor = (productId: string) =>
  `/api/admin/products/${encodeURIComponent(productId)}/relationships`;

export function getAdminProducts(): Promise<Product[]> {
  return request<Product[]>("/api/admin/products");
}

export function getProductRelationships(productId: string): Promise<ProductRelationship[]> {
  return request<ProductRelationship[]>(pathFor(productId));
}

export function replaceProductRelationships(productId: string, relatedProductIds: string[]): Promise<unknown> {
  return request<unknown>(pathFor(productId), {
    method: "PUT",
    body: JSON.stringify({ relatedProductIds }),
  });
}

export function orderedRelatedProductIds(relationships: ProductRelationship[]): string[] {
  return [...relationships]
    .sort((a, b) => a.displayOrder - b.displayOrder || a.id.localeCompare(b.id))
    .map((relationship) => relationship.relatedProductId);
}

export function searchAvailableProducts(
  products: Product[],
  query: string,
  productId: string,
  selectedIds: string[],
): Product[] {
  const term = query.trim().toLocaleLowerCase();
  if (!term) return [];
  const selected = new Set(selectedIds);
  return products.filter((product) =>
    product.id !== productId &&
    !selected.has(product.id) &&
    (product.title.toLocaleLowerCase().includes(term) || product.slug.toLocaleLowerCase().includes(term))
  ).slice(0, 10);
}
