import type { Collection } from "@/types/collection";
import type { ProductType } from "@/types/product-type";
import type { Product, ProductStatus } from "@/types/product";

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

interface ProductCollectionAssignment {
  productId: string;
  collectionId: string;
}

interface ProductTypeAssignment {
  productId: string;
  productTypeId: string;
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

  // Admin endpoints are authentication-free in local development. If the
  // storefront itself is behind Cloudflare Access, forward its credentials.
  for (const name of ["Cf-Access-Jwt-Assertion", "Cookie"]) {
    const value = requestHeaders?.get(name);
    if (value) headers.set(name, value);
  }

  const response = await fetch(apiUrl(path), { headers });

  if (!response.ok) {
    const message = await response.text();
    throw new CatalogApiError(
      `Catalog API ${path} returned ${response.status}${message ? `: ${message}` : ""}`,
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
    // TODO(product-images): hydrate this array when the API exposes image reads.
    images: [],
    // Product files are private purchase assets and must stay out of this response.
    files: [],
  };
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

    return product.status === "active" ? normalizeProduct(product) : null;
  } catch (error) {
    if (error instanceof CatalogApiError && error.status === 404) return null;
    throw error;
  }
}

export async function getCatalog(requestHeaders?: Headers): Promise<CatalogData> {
  const [apiProducts, collections, productTypes] = await Promise.all([
    getJson<ApiProduct[]>("/api/admin/products", requestHeaders),
    getJson<Collection[]>("/api/admin/collections", requestHeaders),
    getJson<ProductType[]>("/api/admin/product-types", requestHeaders),
  ]);

  const activeProducts = apiProducts.filter((product) => product.status === "active");
  const collectionById = new Map(collections.map((item) => [item.id, item]));
  const productTypeById = new Map(productTypes.map((item) => [item.id, item]));

  const products = await Promise.all(
    activeProducts.map(async (product): Promise<CatalogProduct> => {
      const [collectionAssignments, productTypeAssignments] = await Promise.all([
        getJson<ProductCollectionAssignment[]>(
          `/api/admin/products/${product.id}/collections`,
          requestHeaders,
        ),
        getJson<ProductTypeAssignment[]>(
          `/api/admin/products/${product.id}/product-types`,
          requestHeaders,
        ),
      ]);

      const normalizedProduct = normalizeProduct(product);

      return {
        ...normalizedProduct,
        collections: collectionAssignments.flatMap((assignment) => {
          const collection = collectionById.get(assignment.collectionId);
          return collection ? [collection] : [];
        }),
        productTypes: productTypeAssignments.flatMap((assignment) => {
          const productType = productTypeById.get(assignment.productTypeId);
          return productType ? [productType] : [];
        }),
      };
    }),
  );

  return { products, collections, productTypes };
}
