import type { Product } from "@/types/product";

export const products: Product[] = [
  {
    id: "test-the-big-one",
    title: "The Big One",
    slug: "the-big-one",
    description:
      "The Big One family design. A fun digital design perfect for personalizing your favorite projects.",
    price: 3.99,
    status: "active",
    images: [
      {
        id: "test-the-big-one-image",
        url: "/images/the-big-one-1.jpg",
        displayOrder: 0,
      },
      {
        id: "test-the-big-one-image",
        url: "/images/the-big-one-2.jpg",
        displayOrder: 0,
      },
      {
        id: "test-the-big-one-image",
        url: "/images/the-big-one-3.jpg",
        displayOrder: 0,
      },
    ],
    files: [
      {
        id: "test-the-big-one-file",
        fileKey: "test/the-big-one.svg",
        fileName: "the-big-one.svg",
        fileType: "svg",
        displayOrder: 0,
      },
    ],
  },

  {
    id: "test-the-berry-sweet-one",
    title: "The Berry Sweet One",
    slug: "the-berry-sweet-one",
    description:
      "The Berry Sweet One digital design, perfect for creating personalized projects.",
    price: 3.99,
    status: "active",
    images: [
      {
        id: "test-the-berry-sweet-one-image",
        url: "/images/product-placeholder.jpg",
        displayOrder: 0,
      },
    ],
    files: [
      {
        id: "test-the-berry-sweet-one-file",
        fileKey: "test/the-berry-sweet-one.svg",
        fileName: "the-berry-sweet-one.svg",
        fileType: "svg",
        displayOrder: 0,
      },
    ],
  },

  {
    id: "test-sweet-one-boy",
    title: "Sweet One Boy",
    slug: "sweet-one-boy",
    description:
      "Sweet One Boy digital design for creating personalized projects.",
    price: 3.99,
    status: "active",
    images: [
      {
        id: "test-sweet-one-boy-image",
        url: "/images/product-placeholder.jpg",
        displayOrder: 0,
      },
    ],
    files: [
      {
        id: "test-sweet-one-boy-file",
        fileKey: "test/sweet-one-boy.svg",
        fileName: "sweet-one-boy.svg",
        fileType: "svg",
        displayOrder: 0,
      },
    ],
  },

  {
    id: "test-the-big-one-family-members",
    title: "The Big One Family Members",
    slug: "the-big-one-family-members",
    description:
      "Family member designs from The Big One collection.",
    price: 2.99,
    status: "active",
    images: [
      {
        id: "test-the-big-one-family-members-image",
        url: "/images/product-placeholder.jpg",
        displayOrder: 0,
      },
    ],
    files: [
      {
        id: "test-the-big-one-family-members-file",
        fileKey: "test/the-big-one-family-members.svg",
        fileName: "the-big-one-family-members.svg",
        fileType: "svg",
        displayOrder: 0,
      },
    ],
  },

  {
    id: "test-the-big-one-extra-family-members",
    title: "The Big One Extra Family Members",
    slug: "the-big-one-extra-family-members",
    description:
      "Additional family member designs from The Big One collection.",
    price: 2.99,
    status: "active",
    images: [
      {
        id: "test-the-big-one-extra-family-members-image",
        url: "/images/product-placeholder.jpg",
        displayOrder: 0,
      },
    ],
    files: [
      {
        id: "test-the-big-one-extra-family-members-file",
        fileKey: "test/the-big-one-extra-family-members.svg",
        fileName: "the-big-one-extra-family-members.svg",
        fileType: "svg",
        displayOrder: 0,
      },
    ],
  },
];

export const productRelationships: Record<string, string[]> = {
  "test-the-big-one": [
    "test-the-big-one-family-members",
    "test-the-big-one-extra-family-members",
  ],

  "test-the-big-one-family-members": [
    "test-the-big-one",
    "test-the-big-one-extra-family-members",
  ],

  "test-the-big-one-extra-family-members": [
    "test-the-big-one",
    "test-the-big-one-family-members",
  ],
};

export function getRelatedProducts(productId: string): Product[] {
  const relatedIds = productRelationships[productId] ?? [];

  return relatedIds
    .map((relatedId) => products.find((product) => product.id === relatedId))
    .filter((product): product is Product => Boolean(product));
}