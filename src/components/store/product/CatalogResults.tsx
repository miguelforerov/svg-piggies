import type { Product } from "@/types/product";
import { CATALOG_CHANGE_EVENT, navigateCatalog } from "@/lib/catalog-navigation";
import { useEffect, useRef, useState } from "react";
import type { MouseEvent } from "react";
import ProductGrid from "./ProductGrid";
import Pagination from "@/components/store/ui/Pagination";

interface Results {
  products: Product[];
  searchValue: string;
  currentPage: number;
  totalPages: number;
}

interface CatalogResultsProps {
  initialProducts: Product[];
  initialSearchValue: string;
  initialCurrentPage: number;
  initialTotalPages: number;
  initialQuery: string;
}

const CatalogResults = ({
  initialProducts,
  initialSearchValue,
  initialCurrentPage,
  initialTotalPages,
  initialQuery,
}: CatalogResultsProps) => {
  const [query, setQuery] = useState(initialQuery);
  const [results, setResults] = useState<Results>({
    products: initialProducts,
    searchValue: initialSearchValue,
    currentPage: initialCurrentPage,
    totalPages: initialTotalPages,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const loadedQuery = useRef(initialQuery);

  useEffect(() => {
    const syncQuery = () => setQuery(new URLSearchParams(window.location.search).toString());
    const collectionNav = document.querySelector('nav[aria-label="Collections"]');
    const handleCollectionClick = (event: Event) => {
      const mouseEvent = event as globalThis.MouseEvent;
      const target = mouseEvent.target;
      if (
        mouseEvent.button !== 0 || mouseEvent.metaKey || mouseEvent.ctrlKey ||
        mouseEvent.shiftKey || mouseEvent.altKey || !(target instanceof Element)
      ) return;

      const link = target.closest<HTMLAnchorElement>("a[data-catalog-collection]");
      if (!link) return;

      mouseEvent.preventDefault();
      link.classList.remove("hover:text-primary", "duration-300");
      navigateCatalog(link.href);
    };

    window.addEventListener(CATALOG_CHANGE_EVENT, syncQuery);
    window.addEventListener("popstate", syncQuery);
    collectionNav?.addEventListener("click", handleCollectionClick);
    syncQuery();

    return () => {
      window.removeEventListener(CATALOG_CHANGE_EVENT, syncQuery);
      window.removeEventListener("popstate", syncQuery);
      collectionNav?.removeEventListener("click", handleCollectionClick);
    };
  }, []);

  useEffect(() => {
    if (query === loadedQuery.current) {
      setLoading(false);
      setError(false);
      return;
    }

    const controller = new AbortController();
    setLoading(true);
    setError(false);

    const loadResults = async () => {
      try {
        const response = await fetch(`/catalog-results.json${query ? `?${query}` : ""}`, {
          signal: controller.signal,
        });
        if (!response.ok) throw new Error(`Catalog results returned ${response.status}`);

        const nextResults = (await response.json()) as Results;
        setResults(nextResults);
        loadedQuery.current = query;
      } catch (cause) {
        if (!controller.signal.aborted) {
          console.error("Unable to load catalog results:", cause);
          setError(true);
        }
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    };

    void loadResults();
    return () => controller.abort();
  }, [query, retryCount]);

  const handlePageClick = (event: MouseEvent<HTMLAnchorElement>) => {
    if (
      event.button !== 0 || event.metaKey || event.ctrlKey ||
      event.shiftKey || event.altKey
    ) return;

    event.preventDefault();
    navigateCatalog(event.currentTarget.href);
  };

  return (
    <div aria-busy={loading}>
      {error && (
        <p role="alert" className="mb-4 text-red-700">
          Could not update the products.{" "}
          <button type="button" className="underline" onClick={() => setRetryCount((count) => count + 1)}>
            Try again
          </button>
        </p>
      )}
      <ProductGrid initialProducts={results.products} searchValue={results.searchValue} />

      <div className="my-10">
        <Pagination
          currentPage={results.currentPage}
          totalPages={results.totalPages}
          basePath="/products"
          query={query}
          onPageClick={handlePageClick}
        />
      </div>
    </div>
  );
};

export default CatalogResults;
