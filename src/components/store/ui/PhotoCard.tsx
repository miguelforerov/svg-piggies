interface PhotoCardProps {
  image?: { url: string };
  title: string;
}

const PhotoCard = ({ image, title }: PhotoCardProps) => {
  return (
    <div className="flex aspect-square w-full items-center overflow-hidden rounded-default group-hover:border">
      <img
        src={image?.url || "/images/product-placeholder.jpg"}
        width={320}
        height={320}
        alt={title}
        className="h-full w-full object-cover transition duration-300 group-hover:scale-110"
      />
    </div>
  );
};

export default PhotoCard;
