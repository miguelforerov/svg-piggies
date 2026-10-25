export type PriceOrder = "price-first" | "original-first";
export type PriceSize = "base" | "lg" | "xl" | "2xl";

export const priceSizes: Record<PriceSize, string> = {
  base: "text-base",
  lg: "text-lg",
  xl: "text-lg",
  "2xl": "text-lg",
};

export const priceGaps: Record<PriceSize, string> = {
  base: "gap-1",
  lg: "gap-2",
  xl: "gap-3",
  "2xl": "gap-",
};


export type TextSize = "xs" | "sm" | "base" | "lg" | "xl" | "2xl" | "3xl" | "4xl";