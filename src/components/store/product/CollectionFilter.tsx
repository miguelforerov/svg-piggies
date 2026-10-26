import { useEffect, useState, type MouseEvent } from "react";
import type { Collection } from "@/types/collection";
import { Card, CardContent } from "@/components/ui/card";
import {
  CATALOG_CHANGE_EVENT,
  catalogUrl,
  collectionPath,
  getCatalogSearchParams,
  hasActiveCatalogFilters,
  navigateCatalog,
} from "@/lib/catalog-navigation";

interface CollectionFilterProps {
  collections: Collection[];
  initialQuery?: string;
  onNavigate?: () => void;
}

const CollectionFilter = ({ collections, initialQuery = "", onNavigate }: CollectionFilterProps) => {
  const [searchParams, setSearchParams] = useState(() =>
    typeof window === "undefined"
      ? new URLSearchParams(initialQuery)
      : getCatalogSearchParams(window.location),
  );

  useEffect(() => {
    const sync = () => setSearchParams(getCatalogSearchParams(window.location));
    window.addEventListener(CATALOG_CHANGE_EVENT, sync);
    window.addEventListener("popstate", sync);
    return () => {
      window.removeEventListener(CATALOG_CHANGE_EVENT, sync);
      window.removeEventListener("popstate", sync);
    };
  }, []);

  const selectedCollection = searchParams.get("collection");
  const allProductsSelected = !selectedCollection && !hasActiveCatalogFilters(searchParams);

  const collectionLink = (slug?: string) => {
    const params = new URLSearchParams(searchParams);
    params.delete("collection");
    params.delete("c");
    params.delete("pt");
    if (!slug) {
      params.delete("minPrice");
      params.delete("maxPrice");
    }
    params.delete("page");
    return catalogUrl(slug ? collectionPath(slug) : "/products", params);
  };

  const handleClick = (event: MouseEvent<HTMLAnchorElement>) => {
    if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    event.preventDefault();
    navigateCatalog(event.currentTarget.href);
    onNavigate?.();
  };

  return (
    <Card className="bg-muted">
      {collections.length > 0 && (
        <CardContent>
          <nav aria-label="Collections" className="flex flex-col gap-4">
            <div className="font-bold lg:text-lg">Collections</div>
            <ul className="flex flex-col gap-2">
              <li>
                <a
                  data-catalog-collection=""
                  href={collectionLink()}
                  onClick={handleClick}
                  aria-current={allProductsSelected ? "page" : undefined}
                  className={allProductsSelected ? "font-bold cursor-default" : "hover:text-primary duration-300"}
                >
                  All products
                </a>
              </li>
              {collections.map((collection) => {
                const selected = selectedCollection === collection.slug;
                return (
                  <li key={collection.id}>
                    <a
                      data-catalog-collection={collection.slug}
                      href={collectionLink(collection.slug)}
                      onClick={handleClick}
                      aria-current={selected ? "page" : undefined}
                      className={selected ? "font-bold cursor-default" : "hover:text-primary duration-300"}
                    >
                      {collection.name}
                    </a>
                  </li>
                );
              })}
            </ul>
          </nav>
        </CardContent>
      )}
    </Card>
  );
};

export default CollectionFilter;
