import React from "react";
import { cn } from "cn"
import { currency } from "@/types/store";
import type { PriceSize, PriceOrder } from "@/types/design";
import { priceSizes, priceGaps } from "@/types/design";

interface PriceProps {
  price: string;
  originalPrice?: string;
  size?: PriceSize
  priceOrder?: PriceOrder
  className?: string;
}

const Price: React.FC<PriceProps> = ({
  price,
  originalPrice,
  size = "base",
  priceOrder,
  className,
}) => {
  const formatPrice = (value: string) =>
    new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: currency.code,
      currencyDisplay: "narrowSymbol",
    }).format(parseFloat(value));

  const formattedAmount = formatPrice(price);
  const formattedOriginalAmount = originalPrice
    ? formatPrice(originalPrice)
    : null;

  return (
    <div className={`flex items-center ${priceGaps[size]} ${priceOrder === "original-first" ? "flex-row-reverse" : ""}`}>
      <div className={`flex items-center text-price ${priceGaps[size]}`}>
        <span className={`${priceSizes[size]} ${className}`}>{currency.label}</span>
        <span className={`${priceSizes[size]} ${className}`}>{formattedAmount}</span>
      </div>
      {formattedOriginalAmount && (
        <span className={cn(
          "text-muted-foreground line-through",
          size === "base" && "text-xs",
          size === "lg" && "text-sm",
          size === "xl" && "text-base",
          size === "2xl" && "text-base",
        )}>
          {formattedOriginalAmount}
        </span>
      )}
    </div>
  );
};

export default Price;
