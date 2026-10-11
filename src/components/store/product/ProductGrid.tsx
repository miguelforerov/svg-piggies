import React from "react";
import type { Product } from "@/types/product";
import { AddToCart } from "@/components/store/cart/AddToCart";

const ProductGrid = ({
  initialProducts,
  searchValue,
}: {
  initialProducts: Product[];
  searchValue?: string | null;
}) => {
  const currencySymbol = "$";

  const resultsText =
    initialProducts.length === 1 ? "result" : "results";

  return (
    <div className="">
      {searchValue ? (
        <p className="mb-4">
          {initialProducts.length === 0
            ? "There are no products that match "
            : `Showing ${initialProducts.length} ${resultsText} for `}
          <span className="font-bold">
            &quot;{searchValue}&quot;
          </span>
        </p>
      ) : null}

      {initialProducts.length === 0 && (
        <div className="mx-auto pt-5 text-center">
          <h1 className="h2 mb-4">No Product Found!</h1>
          <p>
            We couldn&apos;t find what you were looking for.
          </p>
        </div>
      )}

      <div className="grid sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-8">
        {initialProducts.map((product) => {
          const image = product.images[0];

          return (
            <div
              key={product.id}
              className="relative flex w-full flex-col items-center gap-2 hover:bg-muted rounded-lg"
            >
              <div className="group relative flex w-full flex-col items-center gap-1 lg:gap-2 h-full">
                <div className="relative flex items-center aspect-square overflow-hidden max-w-[20rem] h-auto group-hover:border-border-main group-hover:border rounded-md">
                  <img
                    src={image?.url || "/images/product-placeholder.jpg"}
                    width={320}
                    height={320}
                    alt={product.title}
                    className="rounded-md group-hover:scale-110 transition duration-300"
                  />
                </div>
                <a
                  className="after:absolute after:inset-0 text-center font-medium lg:text-xl line-clamp-2 hover:text-primary duration-300"
                  href={`/products/${product.slug}`}
                >
                  {product.title}
                </a>
                <div className="flex items-center justify-center gap-2">
                  <span className="font-bold text-xl text-price">
                    {currencySymbol} {product.price.toFixed(2)}
                  </span>
                  <span className="text-sm line-through text-muted-foreground">
                    {currencySymbol} {product.price.toFixed(2)}
                  </span>
                </div>
              </div>
              <AddToCart productId={product.id} fullWidth variant="ghost" />
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ProductGrid;
