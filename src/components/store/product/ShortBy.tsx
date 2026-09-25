import { type SortFilterItem, sorting } from "@/lib/constants";
import { Suspense } from "react";
import DropdownMenu from "../../../layouts/functional-components/filter/DropdownMenu";

export type ListItem = SortFilterItem | PathFilterItem;
export type PathFilterItem = { title: string; path: string };

const ShortBy = () => (
  <section className="flex items-center justify-end gap-4">
    <span className="hidden md:block">Sort By</span>
    <Suspense>
      <DropdownMenu list={sorting} />
    </Suspense>
  </section>
);

export default ShortBy;
