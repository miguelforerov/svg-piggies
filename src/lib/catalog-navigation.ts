export const CATALOG_CHANGE_EVENT = "catalog:change";

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
