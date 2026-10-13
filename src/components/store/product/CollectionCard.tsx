interface CollectionCardProps {
  id: string;
  images: { url: string }[];
  title: string;
  url: string;
}

const CollectionCard = ({ id, images, title, url }: CollectionCardProps) => {
  const image = images[0];

  return (
    <article
      id={`collection-${id}`}
      className="group relative flex w-full flex-col items-center gap-2 rounded-lg hover:bg-muted pb-4"
    >
      <div className="flex aspect-square w-full  items-center overflow-hidden rounded-md group-hover:border group-hover:border-border-main">
        <img
          src={image?.url || "/images/product-placeholder.jpg"}
          width={320}
          height={320}
          alt={title}
          className="h-full w-full rounded-md object-cover transition duration-300 group-hover:scale-110"
        />
      </div>
      <a
        className="after:absolute after:inset-0 text-center font-medium text-lg lg:text-2xl hover:text-primary duration-300"
        href={`/collection/${url}`}
      >
        {title}
      </a>
    </article>
  );
};

export default CollectionCard;
