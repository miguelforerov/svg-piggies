import React from "react";
import { currency } from "@/types/store";

interface PriceProps {
  amount: string;
  originalAmount?: string;
}

const Price: React.FC<PriceProps> = ({
  amount,
  originalAmount,
}) => {
  const formatPrice = (value: string) =>
    new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: currency.code,
      currencyDisplay: "narrowSymbol",
    }).format(parseFloat(value));

  const formattedAmount = formatPrice(amount);
  const formattedOriginalAmount = originalAmount
    ? formatPrice(originalAmount)
    : null;



  return (
    <div className="flex items-center gap-4">
      <span className="flex gap-1 text-heading-3 lg:text-heading-2">
        <span className="">{currency.label}</span>
        {formattedAmount}
      </span>
      {formattedOriginalAmount && (
        <span className="text-muted-foreground line-through">
          {formattedOriginalAmount}
        </span>
      )}
    </div>
  );
};

export default Price;
