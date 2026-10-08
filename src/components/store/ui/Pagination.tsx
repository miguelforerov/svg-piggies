import type { MouseEventHandler } from "react";
import {
  Pagination as ShadcnPagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  basePath: string;
  query: string;
  onPageClick?: MouseEventHandler<HTMLAnchorElement>;
}

const Pagination = ({
  currentPage,
  totalPages,
  basePath,
  query,
  onPageClick,
}: PaginationProps) => {
  if (totalPages <= 1) return null;

  const pageLink = (page: number) => {
    const params = new URLSearchParams(query);
    if (page <= 1) params.delete("page");
    else params.set("page", String(page));
    const nextQuery = params.toString();
    return nextQuery ? `${basePath}?${nextQuery}` : basePath;
  };

  const hasPrevious = currentPage > 1;
  const hasNext = currentPage < totalPages;

  return (
    <ShadcnPagination>
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious
            href={hasPrevious ? pageLink(currentPage - 1) : undefined}
            onClick={hasPrevious ? onPageClick : undefined}
            aria-disabled={!hasPrevious}
            className={!hasPrevious ? "pointer-events-none opacity-50" : undefined}
          />
        </PaginationItem>

        {Array.from({ length: totalPages }, (_, index) => index + 1).map((page) => (
          <PaginationItem key={page}>
            <PaginationLink
              href={pageLink(page)}
              isActive={page === currentPage}
              onClick={page === currentPage ? (event) => event.preventDefault() : onPageClick}
            >
              {page}
            </PaginationLink>
          </PaginationItem>
        ))}

        <PaginationItem>
          <PaginationNext
            href={hasNext ? pageLink(currentPage + 1) : undefined}
            onClick={hasNext ? onPageClick : undefined}
            aria-disabled={!hasNext}
            className={!hasNext ? "pointer-events-none opacity-50" : undefined}
          />
        </PaginationItem>
      </PaginationContent>
    </ShadcnPagination>
  );
};

export default Pagination;
