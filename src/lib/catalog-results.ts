import type { CatalogProduct } from "@/lib/catalog";
import { defaultSort, sorting } from "@/lib/constants";
import { PRODUCT_PAGE_SIZES } from "@/lib/product-page-sizes";

export function getCatalogResults(products: CatalogProduct[], params: URLSearchParams) {
  const searchValue = params.get("q")?.trim() ?? "";
  const minPrice = Number(params.get("minPrice") ?? 0);
  const maxPriceParam = params.get("maxPrice");
  const maxPrice = maxPriceParam ? Number(maxPriceParam) : Number.POSITIVE_INFINITY;
  const navCollection = params.get("collection");
  const collectionSlugs = new Set(
    navCollection ? [navCollection] : params.getAll("c").filter(Boolean),
  );
  const productTypeSlugs = new Set(params.getAll("pt").filter(Boolean));
  const { sortKey, reverse } =
    sorting.find((item) => item.slug === (params.get("sort") ?? "")) ?? defaultSort;

  const filteredProducts = products
    .filter((product) => {
      const matchesSearch =
        !searchValue ||
        `${product.title} ${product.description}`
          .toLocaleLowerCase()
          .includes(searchValue.toLocaleLowerCase());
      const matchesPrice = product.price >= minPrice && product.price <= maxPrice;
      const matchesCollection =
        collectionSlugs.size === 0 ||
        product.collections.some((collection) => collectionSlugs.has(collection.slug));
      const matchesProductType =
        productTypeSlugs.size === 0 ||
        product.productTypes.some((productType) => productTypeSlugs.has(productType.slug));

      return matchesSearch && matchesPrice && matchesCollection && matchesProductType;
    })
    .sort((left, right) => {
      if (sortKey !== "PRICE") return 0;

      const comparison = left.price - right.price;
      return reverse ? -comparison : comparison;
    });

  const totalPages = Math.max(1, Math.ceil(filteredProducts.length / PRODUCT_PAGE_SIZES.catalog));
  const requestedPage = Number.parseInt(params.get("page") ?? "1", 10);
  const currentPage = Math.min(
    Number.isFinite(requestedPage) && requestedPage > 0 ? requestedPage : 1,
    totalPages,
  );
  const pageStart = (currentPage - 1) * PRODUCT_PAGE_SIZES.catalog;

  return {
    products: filteredProducts.slice(pageStart, pageStart + PRODUCT_PAGE_SIZES.catalog),
    searchValue,
    currentPage,
    totalPages,
  };
}
