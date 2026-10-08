import type { Collection } from "@/types/collection";
import type { ProductType } from "@/types/product-type";
import { useEffect, useState } from "react";
import { TbFilterX } from "react-icons/tb";
import { CATALOG_CHANGE_EVENT, hasActiveCatalogFilters, navigateCatalog } from "@/lib/catalog-navigation";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Slider } from "@/components/ui/slider";
import { Button } from "@/components/ui/button";

type ProductCount = { id: string; productCount: number };

interface ProductFiltersProps {
  collections: Collection[];
  productTypes: ProductType[];
  maxPriceData: { amount: string; currencyCode: string };
  collectionCounts: ProductCount[];
  productTypeCounts: ProductCount[];
}

const ProductFilters = ({
  collections,
  productTypes,
  maxPriceData,
  collectionCounts,
  productTypeCounts,
}: ProductFiltersProps) => {
  const maxAmount = Math.max(0, parseInt(maxPriceData.amount, 10) || 0);
  const [searchParams, setSearchParams] = useState(
    () => new URLSearchParams(window.location.search),
  );
  const [priceRange, setPriceRange] = useState(() => [
    Number(searchParams.get("minPrice") ?? 0),
    Number(searchParams.get("maxPrice") ?? maxAmount),
  ]);
  const currencySymbol = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: maxPriceData.currencyCode,
  }).formatToParts(0).find((part) => part.type === "currency")?.value ?? "$";

  const selectedCollections = searchParams.getAll("c");
  const selectedNavCollection = searchParams.get("collection");
  const selectedProductTypes = searchParams.getAll("pt");
  const filtersActive = hasActiveCatalogFilters(searchParams) || Boolean(selectedNavCollection);

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
    setPriceRange([
      Number(searchParams.get("minPrice") ?? 0),
      Number(searchParams.get("maxPrice") ?? maxAmount),
    ]);
  }, [searchParams, maxAmount]);

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
      link.classList.toggle("hover:text-primary", !selected);
      link.classList.toggle("duration-300", !selected);
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

  const updatePrice = ([min, max]: readonly number[]) => {
    const params = new URLSearchParams(window.location.search);
    const appliedMin = Number(params.get("minPrice") ?? 0);
    const appliedMax = Number(params.get("maxPrice") ?? maxAmount);
    if (min === appliedMin && max === appliedMax) return;

    if (min === 0) params.delete("minPrice");
    else params.set("minPrice", String(min));
    if (max === maxAmount) params.delete("maxPrice");
    else params.set("maxPrice", String(max));
    params.delete("page");

    const query = params.toString();
    navigateCatalog(query ? `/products?${query}` : "/products");
  };

  return (
    <aside className="flex flex-col gap-10">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between mb-2">
          <div className="font-bold lg:text-lg">Filters</div>
          <Button
            variant="ghost"
            onClick={clearFilters}
            disabled={!filtersActive}
            aria-label="Clear filters"
            title="Clear filters"
            className="flex size-8 items-center justify-center"
          >
            <TbFilterX size={20} aria-hidden="true" />
          </Button>
        </div>
        <div className="font-semibold">Select Price Range</div>
        <Separator />
        <div className="flex flex-col gap-4 mt-2">
          <div className="flex justify-between">
            <span>{currencySymbol}{priceRange[0]} {maxPriceData.currencyCode}</span>
            <span>{currencySymbol}{priceRange[1]} {maxPriceData.currencyCode}</span>
          </div>
          <Slider<number[]>
            aria-label="Price range"
            min={0}
            max={maxAmount}
            minStepsBetweenValues={1}
            value={priceRange}
            onValueChange={setPriceRange}
            onValueCommitted={updatePrice}
            disabled={maxAmount === 0}
          />
        </div>
      </div>

      {collections.length > 0 && (
        <div className="flex flex-col gap-2">
          <div className="font-semibold">Collections</div>
          <Separator />
          <ul className="flex flex-col gap-4 mt-2">
            {collections.map((collection) => (
              <li key={collection.id}>
                <div
                  className={`flex w-full cursor-pointer items-center  justify-between gap-4 ${
                    selectedCollections.includes(collection.slug)
                      ? "font-semibold text-text-primary"
                      : "text-text-gray"
                  }`}
                >
                  <Label htmlFor={`collection-${collection.id}`} className="grow cursor-pointer text-left">
                    {collection.name} ({collectionCounts.find((item) => item.id === collection.id)?.productCount ?? 0})
                  </Label>
                  <Checkbox
                    id={`collection-${collection.id}`}
                    checked={selectedCollections.includes(collection.slug)}
                    onCheckedChange={() => updateFilter("c", collection.slug)}
                  />
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}

      {productTypes.length > 0 && (
        <div className="flex flex-col gap-2">
          <div className="font-semibold">Product Types</div>
          <Separator />
          <ul className="flex flex-col gap-4 mt-2">
            {productTypes.map((productType) => (
              <li key={productType.id}>
                <div
                  className={`flex w-full cursor-pointer items-center justify-between ${
                    selectedProductTypes.includes(productType.slug)
                      ? "font-semibold text-text-primary"
                      : "text-text-gray"
                  }`}
                >
                  <Label htmlFor={`product-type-${productType.id}`} className="grow cursor-pointer">
                    {productType.name} ({productTypeCounts.find((item) => item.id === productType.id)?.productCount ?? 0})
                  </Label>
                  <Checkbox
                    id={`product-type-${productType.id}`}
                    checked={selectedProductTypes.includes(productType.slug)}
                    onCheckedChange={() => updateFilter("pt", productType.slug)}
                  />
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </aside>
  );
};

export default ProductFilters;
