export interface ProductRelationship {
  productId: string;
  relatedProductId: string;
  displayOrder: number;
}

export const productRelationships: ProductRelationship[] = [
  // The Big One → Family Members
  {
    productId: "test-the-big-one",
    relatedProductId: "test-the-big-one-family-members",
    displayOrder: 0,
  },

  // The Big One → Extra Family Members
  {
    productId: "test-the-big-one",
    relatedProductId: "test-the-big-one-extra-family-members",
    displayOrder: 1,
  },

  // Family Members → The Big One
  {
    productId: "test-the-big-one-family-members",
    relatedProductId: "test-the-big-one",
    displayOrder: 0,
  },

  // Family Members → Extra Family Members
  {
    productId: "test-the-big-one-family-members",
    relatedProductId: "test-the-big-one-extra-family-members",
    displayOrder: 1,
  },

  // Extra Family Members → The Big One
  {
    productId: "test-the-big-one-extra-family-members",
    relatedProductId: "test-the-big-one",
    displayOrder: 0,
  },

  // Extra Family Members → Family Members
  {
    productId: "test-the-big-one-extra-family-members",
    relatedProductId: "test-the-big-one-family-members",
    displayOrder: 1,
  },
];