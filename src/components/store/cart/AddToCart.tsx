import { useState } from "react"
import type { ComponentProps } from "react"
import { cn } from "cn"
import { Button } from "@/components/ui/button"
import CartIcon from "@/components/store/icons/Cart"
import { $cartDrawerOpen, addToCart } from "@/cartStore"
import type { Product } from "@/types/product"

interface AddToCartProps {
  product: Pick<Product, "id" | "title" | "price" | "slug" | "images">
  fullWidth?: boolean
  variant?: ComponentProps<typeof Button>["variant"]
  className?: string
}

export function AddToCart({
  product,
  fullWidth,
  variant,
  className,
}: AddToCartProps) {
  const [message, setMessage] = useState<string | null>(null)

  const handleSubmit = () => {
    setMessage(null)

    try {
      addToCart({
        id: product.id,
        productId: product.id,
        variantId: product.id,
        name: product.title,
        price: product.price,
        quantity: 1,
        image: product.images[0]?.url ?? "/images/product-placeholder.jpg",
        slug: product.slug,
      })
      $cartDrawerOpen.set(true)
      setMessage(`${product.title} added to cart`)
    } catch (error) {
      console.error("Failed to add product to cart:", error)
      setMessage("Unable to add product to cart")
    }
  }

  return (
    <>
      <Button
        variant={variant}
        onClick={handleSubmit}
        className={cn(
          "h-12 gap-2 px-6 text-xl font-medium",
          fullWidth && "w-full",
          className
        )}
        size="lg"
      >
        <CartIcon className="size-6" />
        Add to Cart
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
