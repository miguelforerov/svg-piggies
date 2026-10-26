import type { ComponentProps } from "react";
import { TbFilter, TbX } from "react-icons/tb";
import { Button } from "@/components/ui/button";
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarProvider,
  useSidebar,
} from "@/components/ui/sidebar";
import CollectionFilter from "@/components/store/product/CollectionFilter";
import ProductFilters from "@/components/store/product/ProductFilters";
import ShortBy from "@/components/store/product/ShortBy";

interface CatalogSidebarProps {
  filters: ComponentProps<typeof ProductFilters>;
  initialQuery: string;
}

const CatalogSidebarContents = ({ filters, initialQuery }: CatalogSidebarProps) => {
  const { isMobile, openMobile, setOpenMobile, toggleSidebar } = useSidebar();

  return (
    <>
      <div className="flex flex-wrap items-center justify-between gap-3 lg:hidden">
        <Button
          type="button"
          variant="ghost"
          className="gap-2 font-medium"
          aria-label="Refine products"
          aria-expanded={openMobile}
          onClick={toggleSidebar}
        >
          Refine
          <TbFilter aria-hidden="true" />
        </Button>
        <ShortBy initialQuery={initialQuery} />
      </div>
      <Sidebar
        collapsible={isMobile ? "offcanvas" : "none"}
        className={isMobile ? undefined : "hidden lg:flex! h-auto! w-full! bg-transparent!"}
      >
        <SidebarHeader className="flex-row items-center justify-between p-4 lg:hidden">
          <div className="font-bold">Collections &amp; Filters</div>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label="Close collections and filters"
            onClick={() => setOpenMobile(false)}
          >
            <TbX aria-hidden="true" />
          </Button>
        </SidebarHeader>
        <SidebarContent className="gap-4 p-4 lg:overflow-visible lg:p-0">
          <CollectionFilter
            collections={filters.collections}
            initialQuery={initialQuery}
            onNavigate={() => setOpenMobile(false)}
          />
          <ProductFilters {...filters} initialQuery={initialQuery} />
        </SidebarContent>
      </Sidebar>
    </>
  );
};

const CatalogSidebar = ({ filters, initialQuery }: CatalogSidebarProps) => (
  <SidebarProvider mobileBreakpoint={1024} className="block! min-h-0! w-full!">
    <CatalogSidebarContents filters={filters} initialQuery={initialQuery} />
  </SidebarProvider>
);

export default CatalogSidebar;
