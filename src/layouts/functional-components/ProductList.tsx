import type { Product } from "@/types/product";
import { AddToCart } from "./cart/AddToCart";

interface ProductListProps {
  initialProducts: Product[];
  searchValue?: string | null;
}

const ProductList = ({ initialProducts, searchValue }: ProductListProps) => {
  const resultsText = initialProducts.length === 1 ? "result" : "results";

  return (
    <div className="row mx-auto px-4">
      {searchValue ? (
        <p className="mb-4">
          {initialProducts.length === 0
            ? "There are no products that match "
            : `Showing ${initialProducts.length} ${resultsText} for `}
          <span className="font-bold">&quot;{searchValue}&quot;</span>
        </p>
      ) : null}

      {initialProducts.length === 0 && (
        <div className="mx-auto pt-5 text-center">
          <h1 className="h2 mb-4">No Product Found!</h1>
          <p>We couldn&apos;t find what you filtered for. Try filtering again.</p>
        </div>
      )}

      <div className="space-y-10">
        {initialProducts.map((product) => {
          const image = product.images[0];

          return (
            <div className="col-12" key={product.id}>
              <div className="row">
                <div className="col-4">
                  <img
                    src={image?.url || "/images/product-placeholder.jpg"}
                    width={312}
                    height={269}
                    alt={product.title}
                    className="h-[150px] w-[312px] rounded-md border border-border object-cover md:h-[269px]"
                  />
                </div>

                <div className="col-8 py-3 max-md:pt-4">
                  <h2 className="h4 font-bold md:font-normal">
                    <a href={`/products/${product.slug}`}>{product.title}</a>
                  </h2>

                  <div className="mt-2 flex items-center gap-x-2">
                    <span className="text-xs font-bold text-text-light md:text-lg">
                      ${product.price.toFixed(2)} USD
                    </span>
                    {/* TODO(catalog): show the regular price crossed out once
                        sale/compare-at pricing exists in our Product domain. */}
                  </div>

                  <p className="my-4 line-clamp-1 text-text-light max-md:text-xs md:mb-8 md:line-clamp-3">
                    {product.description}
                  </p>
                  <AddToCart
                    productId={product.id}
                    className="btn btn-outline-primary max-md:btn-sm drop-shadow-md"
                  />
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ProductList;
