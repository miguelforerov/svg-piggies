import { useEffect, useState } from "react";
import { CheckIcon, ChevronDownIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { defaultSort, sorting } from "@/lib/constants";
import { CATALOG_CHANGE_EVENT, navigateCatalog } from "@/lib/catalog-navigation";

const ShortBy = () => {
  const [sortSlug, setSortSlug] = useState<string | null>(
    () => new URLSearchParams(window.location.search).get("sort"),
  );
  const activeSort = sorting.find((item) => item.slug === sortSlug) ?? defaultSort;

  useEffect(() => {
    const syncSort = () => setSortSlug(new URLSearchParams(window.location.search).get("sort"));
    window.addEventListener(CATALOG_CHANGE_EVENT, syncSort);
    window.addEventListener("popstate", syncSort);
    return () => {
      window.removeEventListener(CATALOG_CHANGE_EVENT, syncSort);
      window.removeEventListener("popstate", syncSort);
    };
  }, []);

  const selectSort = (slug: string | null) => {
    if (slug === sortSlug) return;

    const params = new URLSearchParams(window.location.search);
    if (slug) params.set("sort", slug);
    else params.delete("sort");
    params.delete("page");

    const query = params.toString();
    navigateCatalog(query ? `/products?${query}` : "/products");
  };

  return (
    <section className="flex items-center justify-end gap-4">
      <span className="hidden md:block text-sm">Sort By</span>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="secondary" />
          }
        >
          {activeSort.title}
          <ChevronDownIcon className="size-4 text-primary" aria-hidden="true" />
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56">
          {sorting.map((item) => (
            <DropdownMenuItem
              key={item.slug ?? "default"}
              onClick={() => selectSort(item.slug)}
              className={item.slug === activeSort.slug ? "font-semibold" : undefined}
            >
              {item.title}
              {item.slug === activeSort.slug && <CheckIcon className="ml-auto" aria-hidden="true" />}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </section>
  );
};

export default ShortBy;
