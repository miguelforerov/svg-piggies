import type { Product } from "@/types/product";
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel";
import RelatedProductCard from "@/components/store/product/RelatedProductCard";

interface RelatedProductsSliderProps {
  products: Product[];
}

const RelatedProductsSlider = ({ products }: RelatedProductsSliderProps) => (
  <Carousel opts={{ align: "start" }} aria-label="Related products" className="w-full">
    <CarouselContent>
      {products.map((product) => (
        <CarouselItem
          key={product.id}
          className="flex-none"
        >
          <RelatedProductCard product={product} />
        </CarouselItem>
      ))}
    </CarouselContent>
    {products.length > 1 && (
      <div className="mt-4 flex justify-center gap-3">
        <CarouselPrevious className="static" aria-label="Previous product" />
        <CarouselNext className="static" aria-label="Next product" />
      </div>
    )}
  </Carousel>
);

export default RelatedProductsSlider;
