import { useState } from "react"
import { LoaderCircle } from "lucide-react"
import Button from "@/layouts/shortcodes/Button"

interface AddToCartProps {
  productId: string
  className?: string
}

export function AddToCart({
  productId,
  className,
}: AddToCartProps) {
  const [pending, setPending] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  const handleSubmit = async () => {
    setPending(true)
    setMessage(null)

    try {
      /*
       * Temporary cart implementation.
       *
       * The real cart/store will be connected later.
       * For now we only verify that the product can be
       * selected without depending on Shopify.
       */

      console.log("Add product to cart:", productId)

      setMessage("Product added to cart")
    } catch (error) {
      console.error("Failed to add product to cart:", error)
      setMessage("Unable to add product to cart")
    } finally {
      setPending(false)
    }
  }

  return (
    <>
      <Button
        onClick={handleSubmit}
        disabled={pending}
        aria-disabled={pending}
        className={className}
      >
        {pending ? (
          <LoaderCircle
            className="animate-spin"
            size={26}
          />
        ) : (
          "Add To Cart"
        )}
      </Button>

      <p
        aria-live="polite"
        className="sr-only"
        role="status"
      >
        {message}
      </p>
    </>
  )
}
