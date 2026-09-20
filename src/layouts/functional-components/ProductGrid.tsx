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

      <div className="grid flex sm:grid-cols-2 md:grid-cols-3 gap-8">
        {initialProducts.map((product) => {
          const image = product.images[0];

          return (
            <div
              key={product.id}
              className="relative group items-center flex flex-col"
            >
              <div className="overflow-hidden relative md:w-[20rem] md:h-[20rem] group-hover:border-border-main group-hover:border rounded-md">
                <img
                  src={
                    image?.url ||
                    "/images/Listing_1_SVG_Piggies_16oz_Christmas_4_Candy_Canes.jpg"
                  }
                  width={320}
                  height={320}
                  alt={product.title}
                  className="mx-auto rounded-md object-cover md:w-[20rem] md:h-[20rem] group-hover:scale-110 transition duration-300"
                />
              </div>

              <div className="z-10 py-2 text-center md:py-4">
                <h2 className="text-base font-medium md:text-xl">
                  <a
                    className="after:absolute after:inset-0 text-primary"
                    href={`/products/${product.slug}`}
                  >
                    {product.title}
                  </a>
                </h2>

                <div className="mt-2 flex flex-wrap items-center justify-center gap-x-2 md:mt-4">
                  <span className="text-base font-bold text-primary md:text-xl">
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