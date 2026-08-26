import config from "@/config/config.json";
import type { Product } from "@/types/product";
import React from "react";

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
    <div className="px-4">
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

      <div className="grid gap-4 sm:grid-cols-2 md:grid-cols-3">
        {initialProducts.map((product) => {
          const image = product.images[0];

          return (
            <div
              key={product.id}
              className="group relative text-center"
            >
              <div className="overflow-hidden md:relative">
                <img
                  src={
                    image?.url ||
                    "/images/product-placeholder.jpg"
                  }
                  width={312}
                  height={269}
                  alt={product.title}
                  className="mx-auto h-[200px] w-full rounded-md border border-border object-cover sm:w-[312px] md:h-[269px]"
                />
              </div>

              <div className="z-20 py-2 text-center md:py-4">
                <h2 className="text-base font-medium md:text-xl">
                  <a
                    className="after:absolute after:inset-0"
                    href={`/products/${product.slug}`}
                  >
                    {product.title}
                  </a>
                </h2>

                <div className="mt-2 flex flex-wrap items-center justify-center gap-x-2 md:mt-4">
                  <span className="text-base font-bold text-text-dark md:text-xl">
                    {currencySymbol} {product.price.toFixed(2)}
                  </span>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ProductGrid;