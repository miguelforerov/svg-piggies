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
    <Carousel opts={{ align: "start" }} aria-label="Collections" className="w-full">
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
        <div className="mt-4 flex justify-center gap-3">
          <CarouselPrevious className="static" aria-label="Previous collection" />
          <CarouselNext className="static" aria-label="Next collection" />
        </div>
      )}
    </Carousel>
  );
};

export default CollectionsSlider;
