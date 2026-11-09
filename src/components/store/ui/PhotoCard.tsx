import { useEffect, useState } from "react";
import { refreshPublicProductImages } from "@/lib/public-product-images";

interface PhotoCardProps {
  image?: { id?: string; url: string; altText?: string | null };
  title: string;
  slug?: string;
}

const placeholder = "/images/product-placeholder.jpg";

const PhotoCard = ({ image, title, slug }: PhotoCardProps) => {
  const [src, setSrc] = useState(image?.url || placeholder);
  useEffect(() => { setSrc(image?.url || placeholder); }, [image?.url]);

  async function recover() {
    if (!slug || !image?.id || src === placeholder) { setSrc(placeholder); return; }
    try {
      const refreshed = (await refreshPublicProductImages(slug)).find((item) => item.id === image.id);
      setSrc(refreshed?.url && refreshed.url !== src ? refreshed.url : placeholder);
    } catch { setSrc(placeholder); }
  }

  return (
    <div className="flex aspect-square w-full items-center overflow-hidden rounded-default group-hover:border">
      <img
        src={src}
        width={320}
        height={320}
        alt={image?.altText ?? title}
        onError={() => void recover()}
        className="h-full w-full object-cover transition duration-300 group-hover:scale-110"
      />
    </div>
  );
};

export default PhotoCard;
