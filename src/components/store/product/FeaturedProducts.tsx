import type { Product } from "@/types/product";
import { Skeleton } from "@/components/ui/skeleton";
import PhotoCard from "@/components/store/ui/PhotoCard";

interface FeaturedProductsProps {
  products: Product[];
}

const FeaturedProducts = ({ products }: FeaturedProductsProps) => {
  if (products.length === 0) {
    return <Skeleton />;
  }

  return (
    <div className="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
      {products.map((product) => {
        const image = product.images.find((item) => item.isPrimary) ?? product.images[0];

        return (
          <article className="group relative text-center" key={product.id}>
            <div className="overflow-hidden rounded-md">
              <PhotoCard image={image} title={product.title} slug={product.slug} />
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
