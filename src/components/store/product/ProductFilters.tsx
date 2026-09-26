import type { Collection } from "@/types/collection";
import type { ProductType } from "@/types/product-type";
import { useEffect, useState } from "react";
import { TbFilterX } from "react-icons/tb";
import { CATALOG_CHANGE_EVENT, hasActiveCatalogFilters, navigateCatalog } from "@/lib/catalog-navigation";
import RangeSlider from "../../rangeSlider/RangeSlider";
import Checkbox from "@/components/store/ui/Checkbox";
import Card from "@/components/store/ui/Card";

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

  const selectedCollections = searchParams.getAll("c");
  const selectedNavCollection = searchParams.get("collection");
  const selectedProductTypes = searchParams.getAll("pt");
  const filtersActive = hasActiveCatalogFilters(searchParams) || Boolean(selectedNavCollection);

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

  useEffect(() => {
    const syncFilters = () => setSearchParams(new URLSearchParams(window.location.search));
    window.addEventListener("popstate", syncFilters);
    window.addEventListener(CATALOG_CHANGE_EVENT, syncFilters);
    return () => {
      window.removeEventListener("popstate", syncFilters);
      window.removeEventListener(CATALOG_CHANGE_EVENT, syncFilters);
    };
  }, []);

  useEffect(() => {
    document.querySelectorAll<HTMLAnchorElement>("a[data-catalog-collection]").forEach((link) => {
      const slug = link.dataset.catalogCollection ?? "";
      const selected = slug
        ? slug === selectedNavCollection
        : !selectedNavCollection && !hasActiveCatalogFilters(searchParams);
      const params = new URLSearchParams(searchParams);
      params.delete("c");
      params.delete("pt");
      if (slug) params.set("collection", slug);
      else {
        params.delete("collection");
        params.delete("minPrice");
        params.delete("maxPrice");
      }
      params.delete("page");

      const query = params.toString();
      link.href = query ? `/products?${query}` : "/products";
      if (selected) link.setAttribute("aria-current", "page");
      else link.removeAttribute("aria-current");
      link.classList.toggle("font-bold", selected);
      link.classList.toggle("text-text-gray", selected);
      link.classList.toggle("cursor-default", selected);
      link.classList.toggle("font-semibold", !selected && !slug);
    });
  }, [searchParams, selectedNavCollection]);

  const updateFilter = (key: "c" | "pt", slug: string) => {
    const newParams = new URLSearchParams(window.location.search);

    newParams.delete("collection");

    const selected = newParams.getAll(key);
    newParams.delete(key);
    for (const value of new Set(selected)) {
      if (value !== slug) newParams.append(key, value);
    }
    if (!selected.includes(slug)) newParams.append(key, slug);
    newParams.delete("page");

    const query = newParams.toString();
    navigateCatalog(query ? `/products?${query}` : "/products");
  };

  const clearFilters = () => {
    const params = new URLSearchParams(window.location.search);
    for (const key of ["collection", "c", "pt", "minPrice", "maxPrice", "page"]) {
      params.delete(key);
    }

    const query = params.toString();
    navigateCatalog(query ? `/products?${query}` : "/products");
  };

  return (
    <div
      id={panelId}
      className={`grid overflow-hidden transition-[grid-template-rows,opacity,visibility] duration-300 ease-in-out motion-reduce:transition-none lg:block lg:visible lg:overflow-visible lg:opacity-100 lg:transition-none ${
        expanded
          ? "visible grid-rows-[1fr] opacity-100"
          : "invisible grid-rows-[0fr] opacity-0"
      }`}
    >
      <div className="min-h-0 overflow-hidden lg:overflow-visible">
        <Card>
        <div className="mb-6 flex items-center justify-between">
          <h2 className="font-bold lg:text-lg">Filters</h2>
          <button
            type="button"
            onClick={clearFilters}
            disabled={!filtersActive}
            aria-label="Clear filters"
            title="Clear filters"
            className="flex size-8 cursor-pointer items-center justify-center rounded-md text-primary transition-colors hover:bg-primary hover:text-white disabled:cursor-default disabled:text-gray-light disabled:bg-transparent"
          >
            <TbFilterX size={20} aria-hidden="true" />
          </button>
        </div>
        <div className="flex flex-col gap-4">
          <div className="font-semibold">Select Price Range</div>
          <hr className="border-border-main" />
          <div className="">
            <RangeSlider maxPriceData={maxPriceData} />
          </div>
        </div>

        {collections.length > 0 && (
          <div>
            <div className="font-semibold">Collections</div>
            <hr className="border-border-main" />
            <ul className="mt-4 space-y-4">
              {collections.map((collection) => (
                <li key={collection.id}>
                  <button
                    type="button"
                    aria-pressed={selectedCollections.includes(collection.slug)}
                    className={`flex w-full cursor-pointer items-center  justify-between gap-4 ${
                      selectedCollections.includes(collection.slug)
                        ? "font-semibold text-text-primary"
                        : "text-text-gray"
                    }`}
                    onClick={() => updateFilter("c", collection.slug)}
                  >
                    <div className="text-left">
                      {collection.name} ({collectionCounts.find((item) => item.id === collection.id)?.productCount ?? 0})
                    </div>
                    <Checkbox checked={selectedCollections.includes(collection.slug)} />
                  </button>
                </li>
              ))}
            </ul>
          </div>
        )}

        {productTypes.length > 0 && (
          <div>
            <div className="font-semibold">Product Types</div>
            <hr className="border-border-main" />
            <ul className="mt-4 space-y-4">
              {productTypes.map((productType) => (
                <li key={productType.id}>
                  <button
                    type="button"
                    aria-pressed={selectedProductTypes.includes(productType.slug)}
                    className={`flex w-full cursor-pointer items-center justify-between ${
                      selectedProductTypes.includes(productType.slug)
                        ? "font-semibold text-text-primary"
                        : "text-text-gray"
                    }`}
                    onClick={() => updateFilter("pt", productType.slug)}
                  >
                    <span>
                      {productType.name} ({productTypeCounts.find((item) => item.id === productType.id)?.productCount ?? 0})
                    </span>
                    <Checkbox checked={selectedProductTypes.includes(productType.slug)} />
                  </button>
                </li>
              ))}
            </ul>
          </div>
        )}
        </Card>
      </div>
    </div>
  );
};

export default ProductFilters;
