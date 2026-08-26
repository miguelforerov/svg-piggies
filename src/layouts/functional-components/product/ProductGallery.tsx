import React, {
  useEffect,
  useRef,
  useState,
  type JSX,
  type MouseEvent,
  type TouchEvent,
} from "react";
import { FiZoomIn, FiX } from "react-icons/fi";
import {
  HiOutlineArrowNarrowLeft,
  HiOutlineArrowNarrowRight,
} from "react-icons/hi";
import type { Swiper as TSwiper } from "swiper";
import "swiper/css";
import "swiper/css/free-mode";
import "swiper/css/navigation";
import "swiper/css/thumbs";
import { FreeMode, Navigation, Thumbs } from "swiper/modules";
import { Swiper, SwiperSlide } from "swiper/react";
import SkeletonProductThumb from "../loadings/skeleton/SkeletonProductThumb";

export interface ImageItem {
  url: string;
  altText: string;
  width: number;
  height: number;
}

interface Position {
  x: number;
  y: number;
}

interface CustomZoomImageProps {
  src: string;
  alt: string;
  width: number;
  height: number;
  onOpenLightbox: () => void;
}

const CustomZoomImage = ({
  src,
  alt,
  width,
  height,
  onOpenLightbox,
}: CustomZoomImageProps): JSX.Element => {
  const [position, setPosition] = useState<Position>({ x: 0.5, y: 0.5 });
  const [showMagnifier, setShowMagnifier] = useState(false);
  const [isTouchDevice, setIsTouchDevice] = useState(false);
  const [touchStartPosition, setTouchStartPosition] =
    useState<Position | null>(null);
  const [touchMoveCount, setTouchMoveCount] = useState(0);
  const imageRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    setIsTouchDevice(
      "ontouchstart" in window || navigator.maxTouchPoints > 0,
    );
  }, []);

  const updatePosition = (clientX: number, clientY: number): void => {
    if (!imageRef.current) return;

    const { left, top, width, height } =
      imageRef.current.getBoundingClientRect();

    const x = Math.max(0, Math.min(1, (clientX - left) / width));
    const y = Math.max(0, Math.min(1, (clientY - top) / height));

    setPosition({ x, y });
  };

  const handleMouseMove = (e: MouseEvent<HTMLDivElement>) => {
    if (isTouchDevice) return;
    updatePosition(e.clientX, e.clientY);
  };

  const handleTouchStart = (e: TouchEvent<HTMLDivElement>) => {
    if (e.touches.length !== 1) return;

    const touch = e.touches[0];

    updatePosition(touch.clientX, touch.clientY);

    setTouchStartPosition({
      x: touch.clientX,
      y: touch.clientY,
    });

    setTouchMoveCount(0);
  };

  const handleTouchMove = (e: TouchEvent<HTMLDivElement>) => {
    if (e.touches.length !== 1) return;

    const touch = e.touches[0];

    updatePosition(touch.clientX, touch.clientY);
    setTouchMoveCount((prev) => prev + 1);
  };

  const handleTouchEnd = () => {
    if (touchMoveCount < 5 && touchStartPosition) {
      onOpenLightbox();
    }

    setTouchStartPosition(null);
  };

  return (
    <div
      className="relative w-full h-full overflow-hidden rounded-md cursor-zoom-in"
      ref={imageRef}
      onMouseEnter={() => !isTouchDevice && setShowMagnifier(true)}
      onMouseLeave={() => !isTouchDevice && setShowMagnifier(false)}
      onMouseMove={handleMouseMove}
      onClick={!isTouchDevice ? onOpenLightbox : undefined}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
    >
      <img
        src={src}
        alt={alt}
        width={width}
        height={height}
        className="w-full h-full object-contain"
        draggable={false}
      />

      {showMagnifier && !isTouchDevice && (
        <div
          className="absolute z-10 flex items-center justify-center bg-white opacity-70 rounded-full p-1 shadow-md"
          style={{
            left: `${position.x * 100}%`,
            top: `${position.y * 100}%`,
            transform: "translate(-50%, -50%)",
            pointerEvents: "none",
            width: "24px",
            height: "24px",
          }}
        >
          <FiZoomIn size={16} />
        </div>
      )}
    </div>
  );
};

interface ProductGalleryProps {
  images: ImageItem[];
}

const ProductGallery = ({ images }: ProductGalleryProps): JSX.Element => {
  const [thumbsSwiper, setThumbsSwiper] = useState<TSwiper | null>(null);
  const [activeIndex, setActiveIndex] = useState<number>(0);
  const [isHovered, setIsHovered] = useState<boolean>(false);
  const [loadingThumb, setLoadingThumb] = useState<boolean>(true);
  const [picUrl, setPicUrl] = useState<string>("");
  const [isTouchDevice, setIsTouchDevice] = useState<boolean>(false);

  const [lightboxOpen, setLightboxOpen] = useState<boolean>(false);
  const [lightboxIndex, setLightboxIndex] = useState<number>(0);

  const prevRef = useRef<HTMLDivElement | null>(null);
  const nextRef = useRef<HTMLDivElement | null>(null);

  const altTextArray = images.map((item) => item.altText);

  useEffect(() => {
    const updateFromURL = () => {
      const params = new URLSearchParams(window.location.search);
      const searchParam = params.get("color");

      if (searchParam) {
        const foundIndex = altTextArray.indexOf(searchParam);
        setActiveIndex(foundIndex !== -1 ? foundIndex : 0);
      }

      setLoadingThumb(false);
    };

    updateFromURL();
    window.addEventListener("popstate", updateFromURL);

    const interval = setInterval(updateFromURL, 500);

    return () => {
      window.removeEventListener("popstate", updateFromURL);
      clearInterval(interval);
    };
  }, [altTextArray]);

  useEffect(() => {
    setIsTouchDevice(
      "ontouchstart" in window || navigator.maxTouchPoints > 0,
    );
  }, []);

  useEffect(() => {
    if (!lightboxOpen) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setLightboxOpen(false);
      }

      if (event.key === "ArrowLeft") {
        setLightboxIndex((prev) =>
          prev === 0 ? filteredImages.length - 1 : prev - 1,
        );
      }

      if (event.key === "ArrowRight") {
        setLightboxIndex((prev) =>
          prev === filteredImages.length - 1 ? 0 : prev + 1,
        );
      }
    };

    document.addEventListener("keydown", handleKeyDown);

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = previousOverflow;
    };
  }, [lightboxOpen]);

  const filteredImages = images.filter(
    (item) => item.altText === altTextArray[activeIndex],
  );

  const handleSlideChange = (swiper: TSwiper): void => {
    setActiveIndex(swiper.activeIndex);
    setPicUrl(filteredImages[swiper.activeIndex]?.url || "");
  };

  const handleThumbSlideClick = (clickedUrl: string): void => {
    const foundIndex = filteredImages.findIndex(
      (item) => item.url === clickedUrl,
    );

    if (foundIndex !== -1) {
      setActiveIndex(foundIndex);
    }
  };

  const openLightbox = (index: number) => {
    setLightboxIndex(index);
    setLightboxOpen(true);
  };

  const closeLightbox = () => {
    setLightboxOpen(false);
  };

  const previousLightboxImage = () => {
    setLightboxIndex((prev) =>
      prev === 0 ? filteredImages.length - 1 : prev - 1,
    );
  };

  const nextLightboxImage = () => {
    setLightboxIndex((prev) =>
      prev === filteredImages.length - 1 ? 0 : prev + 1,
    );
  };

  if (loadingThumb) {
    return <SkeletonProductThumb />;
  }

  return (
    <>
      <div
        className="relative"
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
        <Swiper
          spaceBetween={10}
          thumbs={{ swiper: thumbsSwiper }}
          modules={[FreeMode, Navigation, Thumbs]}
          navigation={{
            prevEl: prevRef.current,
            nextEl: nextRef.current,
          }}
          onSlideChange={handleSlideChange}
          allowTouchMove={!isHovered}
        >
          {filteredImages.map((item, index) => (
            <SwiperSlide key={item.url}>
              <div className="mb-6 border border-border rounded-md max-h-[623px] overflow-hidden">
                <CustomZoomImage
                  src={item.url}
                  alt={item.altText}
                  width={722}
                  height={623}
                  onOpenLightbox={() => openLightbox(index)}
                />
              </div>
            </SwiperSlide>
          ))}

          <div
            className={`hidden lg:block w-full absolute top-1/2 -translate-y-1/2 z-10 px-6 text-text-dark ${
              isHovered
                ? "opacity-100 transition-opacity duration-300 ease-in-out"
                : "opacity-0 transition-opacity duration-300 ease-in-out"
            }`}
          >
            <div
              ref={prevRef}
              className="p-2 lg:p-4 rounded-md bg-body cursor-pointer shadow-sm absolute left-4"
            >
              <HiOutlineArrowNarrowLeft size={24} />
            </div>

            <div
              ref={nextRef}
              className="p-2 lg:p-4 rounded-md bg-body cursor-pointer shadow-sm absolute right-4"
            >
              <HiOutlineArrowNarrowRight size={24} />
            </div>
          </div>
        </Swiper>
      </div>

      <Swiper
        onSwiper={setThumbsSwiper}
        spaceBetween={10}
        slidesPerView={isTouchDevice ? 3.5 : 4}
        freeMode={true}
        watchSlidesProgress={true}
        modules={[FreeMode, Navigation, Thumbs]}
      >
        {filteredImages.map((item) => (
          <SwiperSlide key={item.url}>
            <div
              onClick={() => handleThumbSlideClick(item.url)}
              className={`rounded-md cursor-pointer overflow-hidden ${
                picUrl === item.url
                  ? "border border-primary"
                  : "border border-border"
              }`}
            >
              <img
                src={item.url}
                alt={item.altText}
                width={168}
                height={146}
                className="max-h-[146px]"
                draggable={false}
              />
            </div>
          </SwiperSlide>
        ))}
      </Swiper>

      {lightboxOpen && filteredImages[lightboxIndex] && (
        <div
          className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/90 p-4 md:p-8"
          role="dialog"
          aria-modal="true"
          aria-label="Product image preview"
          onClick={closeLightbox}
        >
          <button
            type="button"
            aria-label="Close image preview"
            onClick={closeLightbox}
            className="absolute top-4 right-4 z-20 flex h-10 w-10 items-center justify-center rounded-full bg-white/90 text-black shadow-md hover:bg-white"
          >
            <FiX size={24} />
          </button>

          {filteredImages.length > 1 && (
            <>
              <button
                type="button"
                aria-label="Previous image"
                onClick={(event) => {
                  event.stopPropagation();
                  previousLightboxImage();
                }}
                className="absolute left-3 md:left-6 z-20 flex h-10 w-10 md:h-12 md:w-12 items-center justify-center rounded-full bg-white/90 text-black shadow-md hover:bg-white"
              >
                <HiOutlineArrowNarrowLeft size={24} />
              </button>

              <button
                type="button"
                aria-label="Next image"
                onClick={(event) => {
                  event.stopPropagation();
                  nextLightboxImage();
                }}
                className="absolute right-3 md:right-6 z-20 flex h-10 w-10 md:h-12 md:w-12 items-center justify-center rounded-full bg-white/90 text-black shadow-md hover:bg-white"
              >
                <HiOutlineArrowNarrowRight size={24} />
              </button>
            </>
          )}

          <div
            className="relative flex h-full w-full items-center justify-center"
            onClick={(event) => event.stopPropagation()}
          >
            <img
              src={filteredImages[lightboxIndex].url}
              alt={filteredImages[lightboxIndex].altText}
              className="max-h-full max-w-full object-contain"
              draggable={false}
            />
          </div>

          {filteredImages.length > 1 && (
            <div className="absolute bottom-4 left-1/2 -translate-x-1/2 rounded-full bg-black/60 px-4 py-2 text-sm text-white">
              {lightboxIndex + 1} / {filteredImages.length}
            </div>
          )}
        </div>
      )}
    </>
  );
};

export default ProductGallery;
