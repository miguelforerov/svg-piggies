import { useId, useState, type ComponentProps } from "react";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";
import { useIsMobile } from "@/hooks/use-mobile";
import CollectionFilter from "@/components/store/product/CollectionFilter";
import ProductFilters from "@/components/store/product/ProductFilters";
import ShortBy from "@/components/store/product/ShortBy";
import CloseIcon from "@/components/store/icons//Close";
import SliderHorizontalIcon from "@/components/store/icons//SliderHorizontal";

interface ProductFilterDrawerProps {
  filters: ComponentProps<typeof ProductFilters>;
  initialQuery: string;
}

const ProductFilterDrawer = ({ filters, initialQuery }: ProductFilterDrawerProps) => {
  const isMobile = useIsMobile(1024);
  const [open, setOpen] = useState(false);
  const triggerId = useId();

  const filterContent = (
    <div className="flex flex-col gap-4">
      <CollectionFilter
        collections={filters.collections}
        initialQuery={initialQuery}
        onNavigate={() => setOpen(false)}
      />
      <ProductFilters {...filters} initialQuery={initialQuery} />
    </div>
  );

  return (
    <>
      <div className="flex flex-wrap items-center justify-between gap-3 lg:hidden">
        {isMobile && (
          <Drawer
            swipeDirection="left"
            open={open}
            onOpenChange={setOpen}
            triggerId={triggerId}
          >
            <DrawerTrigger
              id={triggerId}
              render={<Button type="button" variant="ghost" className="gap-2 font-medium" aria-label="Refine products" />}
            >
              Filters
              <SliderHorizontalIcon aria-hidden="true" />
            </DrawerTrigger>
            <DrawerContent className="">
              <DrawerHeader className="flex-row items-center justify-between p-4">
                <DrawerTitle>Collections &amp; Filters</DrawerTitle>
                <DrawerClose
                  render={<Button type="button" variant="ghost" size="icon" aria-label="Close collections and filters" />}
                >
                  <CloseIcon className="size-[20px]" />
                </DrawerClose>
              </DrawerHeader>
              <div className="min-h-0 flex-1 overflow-y-auto p-4">{filterContent}</div>
              <DrawerFooter className="pt-4">
                <DrawerClose render={<Button type="button" className="w-full" />}>
                  Apply filters
                </DrawerClose>
              </DrawerFooter>
            </DrawerContent>
          </Drawer>
        )}
        <ShortBy initialQuery={initialQuery} />
      </div>
      <div className="hidden lg:block">{!isMobile && filterContent}</div>
    </>
  );
};

export default ProductFilterDrawer;
