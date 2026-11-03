import type { Collection } from "@/types/collection";
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel";
import SkeletonVariant from "@/components/store/ui/SkeletonVariant";
import CollectionCard from "@/components/store/collection/CollectionCard";

export interface CollectionSlide extends Collection {
  productCount: number;
  imageUrl?: string;
}

const CollectionsSlider = ({ collections }: { collections: CollectionSlide[] }) => {
  const collectionsData = collections ?? [];

  if (!collectionsData.length) {
    return (
      <SkeletonVariant variant="card" />
    );
  }

  return (
    <Carousel
      opts={{ align: "start" }}
      aria-label="Collections"
      className={collectionsData.length > 1 ? "w-full px-10 sm:px-0" : "w-full"}
    >
      <CarouselContent className="-ml-2 md:-ml-6">
        {collectionsData.map((item) => (
          <CarouselItem key={item.id} className="basis-full pl-2 sm:basis-1/2 md:basis-1/3 md:pl-6">
            <CollectionCard
              id={item.id}
              images={item.imageUrl ? [{ url: item.imageUrl }] : []}
              title={item.name}
              url={item.slug}
            />
          </CarouselItem>
        ))}
      </CarouselContent>
      {collectionsData.length > 1 && (
        <div className="pointer-events-none absolute inset-x-0 top-1/2 z-10 flex -translate-y-1/2 justify-between sm:pointer-events-auto sm:static sm:mt-4 sm:translate-y-0 sm:justify-center sm:gap-3">
          <CarouselPrevious className="pointer-events-auto static" aria-label="Previous collection" />
          <CarouselNext className="pointer-events-auto static" aria-label="Next collection" />
        </div>
      )}
    </Carousel>
  );
};

export default CollectionsSlider;
