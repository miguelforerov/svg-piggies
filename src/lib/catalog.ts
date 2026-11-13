import type { Collection } from "@/types/collection";
import type { ProductType } from "@/types/product-type";
import type { Product, ProductStatus } from "@/types/product";
import { normalizePublicImages, type PublicImageRecord } from "@/lib/product-image-mapping";

interface ApiProduct {
  id: string;
  title: string;
  slug: string;
  description: string;
  price: string;
  status: ProductStatus;
  createdAt?: string;
  updatedAt?: string;
}

export class CatalogApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "CatalogApiError";
  }
}

interface PublicCatalogResponse {
  products: Array<{ product: ApiProduct; collectionIds: string[]; productTypeIds: string[] }>;
  collections: Collection[];
  productTypes: ProductType[];
}

export interface CatalogProduct extends Product {
  collections: Collection[];
  productTypes: ProductType[];
  createdAt?: string;
  updatedAt?: string;
}

export interface CatalogData {
  products: CatalogProduct[];
  collections: Collection[];
  productTypes: ProductType[];
}

function apiUrl(path: string): string {
  const baseUrl = import.meta.env.PUBLIC_API_BASE_URL?.replace(/\/$/, "");

  if (!baseUrl) {
    throw new Error("PUBLIC_API_BASE_URL is not configured");
  }

  return `${baseUrl}${path}`;
}

async function getJson<T>(path: string, requestHeaders?: Headers): Promise<T> {
  const headers = new Headers({ Accept: "application/json" });

  // Forward storefront credentials when the site itself is behind Access.
  for (const name of ["Cf-Access-Jwt-Assertion", "Cookie"]) {
    const value = requestHeaders?.get(name);
    if (value) headers.set(name, value);
  }

  const response = await fetch(apiUrl(path), { headers, cache: "no-store" });

  if (!response.ok) {
    throw new CatalogApiError(
      `Catalog API request failed with status ${response.status}`,
      response.status,
    );
  }

  return response.json() as Promise<T>;
}

function normalizeProduct(product: ApiProduct): Product {
  const price = Number(product.price);
  if (!Number.isFinite(price)) {
    throw new Error(`Product ${product.id} has an invalid price: ${product.price}`);
  }

  return {
    ...product,
    price,
    images: [],
    // Product files are private purchase assets and must stay out of this response.
    files: [],
  };
}

async function getPublicImages(slug: string, requestHeaders?: Headers): Promise<Product["images"]> {
  try {
    const records = await getJson<PublicImageRecord[]>(`/api/products/${encodeURIComponent(slug)}/images`, requestHeaders);
    return normalizePublicImages(records);
  } catch (error) {
    if (error instanceof CatalogApiError && error.status === 502) return [];
    throw error;
  }
}

export async function getProductBySlug(
  slug: string,
  requestHeaders?: Headers,
): Promise<Product | null> {
  try {
    const product = await getJson<ApiProduct>(
      `/api/products/${encodeURIComponent(slug)}`,
      requestHeaders,
    );

    if (product.status !== "active") return null;
    return { ...normalizeProduct(product), images: await getPublicImages(product.slug, requestHeaders) };
  } catch (error) {
    if (error instanceof CatalogApiError && error.status === 404) return null;
    throw error;
  }
}

export async function getCatalog(requestHeaders?: Headers): Promise<CatalogData> {
  const { products: catalogEntries, collections, productTypes } =
    await getJson<PublicCatalogResponse>("/api/catalog", requestHeaders);
  const collectionById = new Map(collections.map((item) => [item.id, item]));
  const productTypeById = new Map(productTypes.map((item) => [item.id, item]));

  const products = await Promise.all(
    catalogEntries.map(async ({ product, collectionIds, productTypeIds }): Promise<CatalogProduct> => {
      const normalizedProduct = normalizeProduct(product);
      normalizedProduct.images = await getPublicImages(product.slug, requestHeaders);

      return {
        ...normalizedProduct,
        collections: collectionIds.flatMap((id) => {
          const collection = collectionById.get(id);
          return collection ? [collection] : [];
        }),
        productTypes: productTypeIds.flatMap((id) => {
          const productType = productTypeById.get(id);
          return productType ? [productType] : [];
        }),
      };
    }),
  );

  return { products, collections, productTypes };
}

export async function getRelatedProducts(
  product: Product,
  requestHeaders?: Headers,
): Promise<Product[]> {
  const relatedIds = await getJson<string[]>(
    `/api/products/${encodeURIComponent(product.slug)}/relationships`,
    requestHeaders,
  );
  if (relatedIds.length === 0) return [];

  const { products } = await getCatalog(requestHeaders);
  const productsById = new Map(products.map((item) => [item.id, item]));
  return relatedIds.flatMap((id) => {
    const related = productsById.get(id);
    return related && related.id !== product.id ? [related] : [];
  });
}
