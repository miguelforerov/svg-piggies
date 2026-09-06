import { EllipsisVertical, Package } from "lucide-react"
import { Button } from "@/components/ui/button"
interface Product {
  id: string
  title: string
  price: string
  status: "draft" | "active" | "archived"
}

interface ProductCardProps {
  product: Product
}

export function ProductCard({
  product,
}: ProductCardProps) {
  return (
    <div className="rounded-xl border bg-card p-4 text-card-foreground shadow-sm">
      <div className="flex gap-4">
        {/* Thumbnail */}
        <div className="flex size-24 shrink-0 items-center justify-center rounded-lg bg-muted">
          <Package className="size-8 text-muted-foreground" />
        </div>

        {/* Product information */}
        <div className="min-w-0 flex-1">
          <h3 className="truncate font-medium">
            {product.title}
          </h3>

          <p className="mt-2 text-sm font-medium">
            ${product.price}
          </p>

          <span
            className={[
              "mt-2 inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium",
              product.status === "active" &&
                "bg-primary/10 text-primary",
              product.status === "draft" &&
                "bg-muted text-muted-foreground",
              product.status === "archived" &&
                "bg-destructive/10 text-destructive",
            ]
              .filter(Boolean)
              .join(" ")}
          >
            {product.status}
          </span>
        </div>
      </div>

      {/* Actions */}
      <div className="mt-4 flex items-center justify-end gap-2 border-t pt-3">
        <Button variant="link">
          View
        </Button>
      </div>
    </div>
  )
}