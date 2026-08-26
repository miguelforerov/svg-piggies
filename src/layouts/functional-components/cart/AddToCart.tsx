import React, { useState } from "react";
import { BiLoaderAlt } from "react-icons/bi";

interface AddToCartProps {
  productId: string;
  stylesClass?: string;
}

export function AddToCart({
  productId,
  stylesClass = "btn btn-primary",
}: AddToCartProps) {
  const [pending, setPending] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  const handleSubmit = async () => {
    setPending(true);
    setMessage(null);

    try {
      /*
       * Temporary cart implementation.
       *
       * The real cart/store will be connected later.
       * For now we only verify that the product can be
       * selected without depending on Shopify.
       */

      console.log("Add product to cart:", productId);

      setMessage("Product added to cart");
    } catch (error) {
      console.error("Failed to add product to cart:", error);
      setMessage("Unable to add product to cart");
    } finally {
      setPending(false);
    }
  };

  return (
    <>
      <button
        type="button"
        onClick={handleSubmit}
        disabled={pending}
        aria-label="Add to cart"
        aria-disabled={pending}
        className={`${stylesClass} ${
          pending ? "cursor-not-allowed opacity-70" : ""
        }`}
      >
        {pending ? (
          <BiLoaderAlt
            className="animate-spin"
            size={26}
          />
        ) : (
          "Add To Cart"
        )}
      </button>

      <p
        aria-live="polite"
        className="sr-only"
        role="status"
      >
        {message}
      </p>
    </>
  );
}