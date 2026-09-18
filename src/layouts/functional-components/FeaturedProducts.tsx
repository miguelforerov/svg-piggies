import type { Product } from "@/types/product";
import SkeletonFeaturedProducts from "./loadings/skeleton/SkeletonFeaturedProducts";

interface FeaturedProductsProps {
  products: Product[];
}

const FeaturedProducts = ({ products }: FeaturedProductsProps) => {
  if (products.length === 0) {
    return <SkeletonFeaturedProducts />;
  }

  return (
    <div className="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
      {products.map((product) => {
        const image = product.images[0];

        return (
          <article className="group relative text-center" key={product.id}>
            <div className="overflow-hidden rounded-md">
              <img
                src={image?.url || "/images/Listing_1_SVG_Piggies_16oz_Christmas_4_Candy_Canes.jpg"}
                width={312}
                height={269}
                alt={product.title}
                className="h-[150px] w-full rounded-md border border-border object-cover transition duration-300 group-hover:scale-105 md:h-[269px]"
              />
            </div>

            <div className="py-4 text-center">
              <h3 className="text-base font-medium md:text-xl">
                <a
                  className="after:absolute after:inset-0"
                  href={`/products/${product.slug}`}
                >
                  {product.title}
                </a>
              </h3>
              <p className="mt-2 font-bold text-text-primary md:text-lg">
                ${product.price.toFixed(2)}
              </p>
              {/* TODO(catalog): show sale/compare-at price when supported. */}
            </div>
          </article>
        );
      })}
    </div>
  );
};

export default FeaturedProducts;
