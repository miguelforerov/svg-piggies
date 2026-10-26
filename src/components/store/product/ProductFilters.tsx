import type { Collection } from "@/types/collection";
import type { ProductType } from "@/types/product-type";
import { useEffect, useId, useState } from "react";
import { TbFilterX } from "react-icons/tb";
import { CATALOG_CHANGE_EVENT, catalogUrl, getCatalogSearchParams, hasActiveCatalogFilters, navigateCatalog } from "@/lib/catalog-navigation";
import { Checkbox } from "@/components/ui/checkbox";
import { Separator } from "@/components/ui/separator";
import { Slider } from "@/components/ui/slider";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";


type ProductCount = { id: string; productCount: number };

interface ProductFiltersProps {
  initialQuery?: string;
  collections: Collection[];
  productTypes: ProductType[];
  maxPriceData: { amount: string; currencyCode: string };
  collectionCounts: ProductCount[];
  productTypeCounts: ProductCount[];
}

const ProductFilters = ({
  initialQuery = "",
  collections,
  productTypes,
  maxPriceData,
  collectionCounts,
  productTypeCounts,
}: ProductFiltersProps) => {
  const maxAmount = Math.max(0, parseInt(maxPriceData.amount, 10) || 0);
  const [searchParams, setSearchParams] = useState(
    () => typeof window === "undefined"
      ? new URLSearchParams(initialQuery)
      : getCatalogSearchParams(window.location),
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
    const syncFilters = () => setSearchParams(getCatalogSearchParams(window.location));
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

  const updateFilter = (key: "c" | "pt", slug: string) => {
    const newParams = getCatalogSearchParams(window.location);

    newParams.delete("collection");

    const selected = newParams.getAll(key);
    newParams.delete(key);
    for (const value of new Set(selected)) {
      if (value !== slug) newParams.append(key, value);
    }
    if (!selected.includes(slug)) newParams.append(key, slug);
    newParams.delete("page");

    navigateCatalog(catalogUrl("/products", newParams));
  };

  const clearFilters = () => {
    const params = new URLSearchParams(window.location.search);
    for (const key of ["collection", "c", "pt", "minPrice", "maxPrice", "page"]) {
      params.delete(key);
    }

    navigateCatalog(catalogUrl("/products", params));
  };

  const updatePrice = ([min, max]: readonly number[]) => {
    const params = getCatalogSearchParams(window.location);
    const appliedMin = Number(params.get("minPrice") ?? 0);
    const appliedMax = Number(params.get("maxPrice") ?? maxAmount);
    if (min === appliedMin && max === appliedMax) return;

    if (min === 0) params.delete("minPrice");
    else params.set("minPrice", String(min));
    if (max === maxAmount) params.delete("maxPrice");
    else params.set("maxPrice", String(max));
    params.delete("page");

    navigateCatalog(catalogUrl(window.location.pathname, params));
  };

  return (
    <Card className="bg-muted">
      <CardContent className="flex flex-col gap-6">
        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between mb-2">
            <div className="text-heading-6 lg:text-heading-4">Filters</div>
            <Button
              variant="ghost"
              onClick={clearFilters}
              disabled={!filtersActive}
              aria-label="Clear filters"
              title="Clear filters"
              size="icon"
              className="flex items-center justify-center"
            >
              <TbFilterX aria-hidden="true" />
            </Button>
          </div>
          <SubtitleFilter label="Select Price Range" />
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
            <SubtitleFilter label="Collections" />
            <Separator />
            <ul className="flex flex-col gap-2">
              {collections.map((collection) => (
                <ItemFilter
                  key={collection.id}
                  itemId={`collection-${collection.id}`}
                  label={`${collection.name} (${collectionCounts.find((item) => item.id === collection.id)?.productCount ?? 0})`}
                  selected={selectedCollections.includes(collection.slug)}
                  onChange={() => updateFilter("c", collection.slug)}
                />
              ))}
            </ul>
          </div>
        )}

        {productTypes.length > 0 && (
          <div className="flex flex-col gap-2">
            <SubtitleFilter label="Product Types" />
            <Separator />
            <ul className="flex flex-col gap-2">
              {productTypes.map((productType) => (
                <ItemFilter
                  key={productType.id}
                  itemId={`product-type-${productType.id}`}
                  label={`${productType.name} (${productTypeCounts.find((item) => item.id === productType.id)?.productCount ?? 0})`}
                  selected={selectedProductTypes.includes(productType.slug)}
                  onChange={() => updateFilter("pt", productType.slug)}
                />
              ))}
            </ul>
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default ProductFilters;

interface ItemFilterProps {
  label: string;
  selected?: boolean;
  checkbox?: boolean;
  itemId?: string;
  onChange?: () => void;
}

const ItemFilter = ({
  label,
  selected = false,
  checkbox = true,
  itemId,
  onChange
}: ItemFilterProps) => {
  const generatedId = useId();
  const id = itemId ?? generatedId;

  return (
    <li className={`flex w-full items-center justify-between gap-4 ${selected ? "font-bold" : ""}`}>
      <span className="grow text-left">{label}</span>
      {checkbox && (
        <Checkbox id={id} aria-label={label} checked={selected} onCheckedChange={onChange} />
      )}
    </li>
  );
};

interface SubtitleFilterProps {
  label: string;
}

const SubtitleFilter = ({ label }: SubtitleFilterProps) => {
  return (
    <div className="font-medium text-sm lg:text-base">
      {label}
    </div>
  );
};
