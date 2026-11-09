import type { ProductImage } from "@/types/product";
import { normalizePublicImages, type PublicImageRecord } from "@/lib/product-image-mapping";

export async function refreshPublicProductImages(slug: string): Promise<ProductImage[]> {
  const response = await fetch(`/api/products/${encodeURIComponent(slug)}/images`, { cache: "no-store" });
  if (!response.ok) throw new Error("Could not refresh product images");
  return normalizePublicImages(await response.json() as PublicImageRecord[]);
}
