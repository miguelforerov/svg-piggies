export type ProductStatus = "draft" | "active" | "archived";

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