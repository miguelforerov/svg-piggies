import type { Collection } from "@/types/collection";
import type { ProductType } from "@/types/product-type";
import { useEffect, useState } from "react";
import { BsCheckLg } from "react-icons/bs";
import RangeSlider from "../rangeSlider/RangeSlider";

type ProductCount = { id: string; productCount: number };

interface ProductFiltersProps {
  toggleId: string;
  panelId: string;
  collections: Collection[];
  productTypes: ProductType[];
  maxPriceData: { amount: string; currencyCode: string };
  collectionCounts: ProductCount[];
  productTypeCounts: ProductCount[];
}

const ProductFilters = ({
  toggleId,
  panelId,
  collections,
  productTypes,
  maxPriceData,
  collectionCounts,
  productTypeCounts,
}: ProductFiltersProps) => {
  const [expanded, setExpanded] = useState(false);
  const [searchParams, setSearchParams] = useState(
    () => new URLSearchParams(window.location.search),
  );

  const selectedCollection = searchParams.get("c");
  const selectedProductType = searchParams.get("pt");

  useEffect(() => {
    const toggle = document.getElementById(toggleId);
    if (!toggle) return;

    const handleToggle = () => setExpanded((current) => !current);
    toggle.addEventListener("click", handleToggle);
    return () => toggle.removeEventListener("click", handleToggle);
  }, [toggleId]);

  useEffect(() => {
    document.getElementById(toggleId)?.setAttribute("aria-expanded", String(expanded));
  }, [expanded, toggleId]);

  const updateFilter = (key: "c" | "pt", slug: string) => {
    const newParams = new URLSearchParams(searchParams);

    if (newParams.get(key) === slug) {
      newParams.delete(key);
    } else {
      newParams.set(key, slug);
    }
    newParams.delete("page");

    const query = newParams.toString();
    window.location.href = query ? `/products?${query}` : "/products";
    setSearchParams(newParams);
  };

  return (
    <div className="w-full">
      <div
        id={panelId}
        className={`${expanded ? "block" : "hidden"} mt-4 rounded-md border border-border-main p-4 lg:mt-0 lg:block lg:rounded-none lg:border-0 lg:p-0`}
      >
        <div>
          <h5 className="mb-2 lg:text-xl">Select Price Range</h5>
          <hr className="border-border-main" />
          <div className="pt-4">
            <RangeSlider maxPriceData={maxPriceData} />
          </div>
        </div>

        {collections.length > 0 && (
          <div>
            <h5 className="mb-2 mt-4 lg:mt-6 lg:text-xl">Collections</h5>
            <hr className="border-border-main" />
            <ul className="mt-4 space-y-4">
              {collections.map((collection) => (
                <li key={collection.id}>
                  <button
                    type="button"
                    className={`flex w-full cursor-pointer items-center justify-between gap-4 ${
                      selectedCollection === collection.slug
                        ? "font-semibold text-text-primary"
                        : "text-text-gray"
                    }`}
                    onClick={() => updateFilter("c", collection.slug)}
                  >
                    <span>
                      {collection.name} ({collectionCounts.find((item) => item.id === collection.id)?.productCount ?? 0})
                    </span>
                    <span className="flex h-4 w-4 items-center justify-center rounded-sm border border-border-main">
                      {selectedCollection === collection.slug && <BsCheckLg size={16} />}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          </div>
        )}

        {productTypes.length > 0 && (
          <div>
            <h5 className="mb-2 mt-8 lg:mt-10 lg:text-xl">Product Types</h5>
            <hr className="border-border-main" />
            <ul className="mt-4 space-y-4">
              {productTypes.map((productType) => (
                <li key={productType.id}>
                  <button
                    type="button"
                    className={`flex w-full cursor-pointer items-center justify-between ${
                      selectedProductType === productType.slug
                        ? "font-semibold text-text-primary"
                        : "text-text-gray"
                    }`}
                    onClick={() => updateFilter("pt", productType.slug)}
                  >
                    <span>
                      {productType.name} ({productTypeCounts.find((item) => item.id === productType.id)?.productCount ?? 0})
                    </span>
                    <span className="flex h-4 w-4 items-center justify-center rounded-sm border border-border-main">
                      {selectedProductType === productType.slug && <BsCheckLg size={16} />}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </div>
  );
};

export default ProductFilters;
