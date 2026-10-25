import React from "react";
import type { Product } from "@/types/product";
import { AddToCart } from "@/components/store/cart/AddToCart";
import PhotoCard from "@/components/store/ui/PhotoCard"
import Price from "@/components/store/product/Price"

const RelatedProductCard = ({
  product,
}: {
  product: Product;
  searchValue?: string | null;
}) => {

  const image = product.images[0];

  return (
    <div className="w-[10rem] h-[16.5rem]">
      <div
        key={product.id}
        className="relative flex w-full h-full flex-col items-center gap-2 hover:bg-muted rounded-lg duration-300"
      >
        <div className="group relative flex w-full flex-col items-center gap-1 h-full">
          <PhotoCard image={image} title={product.title} />
          <a
            className="after:absolute after:inset-0 text-xs text-center font-medium line-clamp-2 hover:text-primary duration-300"
            href={`/products/${product.slug}`}
          >
            {product.title}
          </a>
          <div className="flex items-center justify-center gap-2">
            <Price
              originalPrice={product.price.toFixed(2)}
              price={product.price.toFixed(2)}
            />
          </div>
        </div>
        <AddToCart product={product} fullWidth variant="secondary" size="sm" />
      </div>
    </div>
  );
};

export default RelatedProductCard;
