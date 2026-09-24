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
              className="group relative flex w-full flex-col items-center"
            >
              <div className="relative flex items-center aspect-square overflow-hidden max-w-[20rem] h-auto group-hover:border-border-main group-hover:border rounded-md">
                <img
                  src={
                    image?.url ||
                    "/images/product-placeholder.jpg"
                  }
                  width={320}
                  height={320}
                  alt={product.title}
                  className="rounded-md group-hover:scale-110 transition duration-300"
                />
              </div>

              <div className="z-10 py-2 text-center md:py-4">
                <h2 className="font-medium text-base lg:text-xl">
                  <a
                    className="after:absolute after:inset-0 text-text-primary group-hover:text-text-primary/80"
                    href={`/products/${product.slug}`}
                  >
                    {product.title}
                  </a>
                </h2>

                <div className="mt-2 flex flex-wrap items-center justify-center gap-x-2 md:mt-4">
                  <span className="font-bold text-text-primary md:text-xl group-hover:text-text-primary/80">
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
