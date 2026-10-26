export const CATALOG_CHANGE_EVENT = "catalog:change";

const COLLECTION_PATH = /^\/products\/collections\/([^/]+)\/?$/;

export function collectionPath(slug: string) {
  return `/products/collections/${encodeURIComponent(slug)}`;
}

export function getCatalogSearchParams(location: Pick<Location, "pathname" | "search">) {
  const params = new URLSearchParams(location.search);
  const match = location.pathname.match(COLLECTION_PATH);
  if (match) params.set("collection", decodeURIComponent(match[1]));
  return params;
}

export function catalogUrl(pathname: string, params: URLSearchParams) {
  const urlParams = new URLSearchParams(params);
  if (COLLECTION_PATH.test(pathname)) urlParams.delete("collection");
  const query = urlParams.toString();
  return query ? `${pathname}?${query}` : pathname;
}

export function hasActiveCatalogFilters(params: URLSearchParams) {
  return (
    params.getAll("c").some(Boolean) ||
    params.getAll("pt").some(Boolean) ||
    params.has("minPrice") ||
    params.has("maxPrice")
  );
}

export function navigateCatalog(url: string) {
  window.history.pushState(null, "", url);
  window.dispatchEvent(new Event(CATALOG_CHANGE_EVENT));
}
