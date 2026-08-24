import { addToCart } from "@/cartStore";

export default function TestAddToCart() {
  const addVariant = (
    variantId: string,
    variantName: string,
  ) => {
    addToCart({
      id: `${variantId}-test`,
      productId: "family-members",
      variantId,
      name: "Family Members",
      variantName,
      price: 2.99,
      quantity: 1,
      image: "/images/product-placeholder.jpg",
      slug: "family-members",
    });
  };

  return (
    <div className="flex gap-4">
      <button
        className="btn btn-primary"
        onClick={() => addVariant("mom", "Mom")}
      >
        Add Mom
      </button>

      <button
        className="btn btn-primary"
        onClick={() => addVariant("dad", "Dad")}
      >
        Add Dad
      </button>

      <button
        className="btn btn-primary"
        onClick={() => addVariant("sister", "Sister")}
      >
        Add Sister
      </button>
    </div>
  );
}