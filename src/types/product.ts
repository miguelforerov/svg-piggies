export const PRODUCT_STATUSES = ["draft", "active", "archived"] as const;

export type ProductStatus = (typeof PRODUCT_STATUSES)[number];

export interface ProductImage {
  id: string;
  url: string;
  displayOrder: number;
}

export interface ProductFile {
  id: string;
  fileKey: string;
  fileName: string;
  fileType: string;
  displayOrder: number;
}

export interface Product {
  id: string;
  title: string;
  slug: string;
  description: string;
  price: number;
  status: ProductStatus;
  images: ProductImage[];
  files: ProductFile[];
}
