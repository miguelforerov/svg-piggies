import { layoutView } from "@/cartStore";
import type { Product } from "@/types/product";
import { useStore } from "@nanostores/react";
import ProductGrid from "./ProductGrid";
import ProductList from "./ProductList";

interface ProductLayoutViewsProps {
  initialProducts: Product[];
  searchValue?: string | null;
}

const ProductLayoutViews = ({
  initialProducts,
  searchValue,
}: ProductLayoutViewsProps) => {
  const layout = useStore(layoutView);

  return (
    <div className="col-12 lg:col-9">
      {layout === "list" ? (
        <ProductList initialProducts={initialProducts} searchValue={searchValue} />
      ) : (
        <ProductGrid initialProducts={initialProducts} searchValue={searchValue} />
      )}
    </div>
  );
};

export default ProductLayoutViews;
