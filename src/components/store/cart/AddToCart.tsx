import { useState } from "react"
import { LoaderCircle } from "lucide-react"
import type { ComponentProps } from "react"
import { cn } from "cn"
import { Button } from "@/components/ui/button"
import Cart from "@/components/store/icons/Cart"

interface AddToCartProps {
  productId: string
  fullWidth?: boolean
  variant?: ComponentProps<typeof Button>["variant"]
  className?: string
}

export function AddToCart({
  productId,
  fullWidth,
  variant,
  className,
}: AddToCartProps) {
  const [pending, setPending] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  const handleSubmit = async () => {
    setPending(true)
    setMessage(null)

    try {
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
        variant={variant}
        onClick={handleSubmit}
        disabled={pending}
        aria-disabled={pending}
        className={cn(
          "h-12 gap-2 px-6 text-xl font-medium",
          fullWidth && "w-full",
          className
        )}
        size="lg"
      >
        {pending ? (
          <LoaderCircle
            className="animate-spin"
            size={26}
          />
        ) : (
          <>
            <Cart data-icon="inline-start" className="size-6" />
            Add to Cart
          </>
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
