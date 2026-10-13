import PhotoCard from "@/components/store/ui/PhotoCard"

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
      className="group relative flex w-full flex-col items-center gap-2 rounded-default hover:bg-muted pb-4"
    >
      <PhotoCard image={image} title={title} />
      <a
        className="after:absolute after:inset-0 text-center text-heading-4 lg:text-heading-3 hover:text-primary duration-300"
        href={`/collection/${url}`}
      >
        {title}
      </a>
    </article>
  );
};

export default CollectionCard;
