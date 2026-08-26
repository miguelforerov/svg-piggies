import { products } from "@/data/products";
import { productRelationships } from "@/data/productRelationships";
import type { Product } from "@/types/product";

export function getProducts(): Product[] {
  return products.filter((product) => product.status === "active");
}

export function getProductBySlug(slug: string): Product | undefined {
  const product = products.find(
    (item) => item.slug === slug && item.status === "active",
  );

  if (!product) {
    return undefined;
  }

  const relationships = productRelationships
    .filter((relationship) => relationship.productId === product.id)
    .sort((a, b) => a.displayOrder - b.displayOrder);

  const relatedProducts = relationships
    .map((relationship) =>
      products.find(
        (relatedProduct) =>
          relatedProduct.id === relationship.relatedProductId &&
          relatedProduct.status === "active",
      ),
    )
    .filter((relatedProduct): relatedProduct is Product => Boolean(relatedProduct));

  return {
    ...product,
    relatedProducts,
  };
}