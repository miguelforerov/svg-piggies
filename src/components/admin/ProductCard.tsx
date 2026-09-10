import { Package } from "lucide-react"
import type { ProductStatus } from "@/types/product"
import { ProductStatusBadge } from "@/components/admin/ProductStatusBadge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

interface Product {
  id: string
  title: string
  price: string
  status: ProductStatus
}

interface ProductCardProps {
  product: Product
}

export function ProductCard({
  product,
}: ProductCardProps) {
  return (
    <Card className="shadow-sm">
      <CardContent>
        <div className="flex gap-4">
          {/* Thumbnail */}
          <div className="flex size-24 shrink-0 items-center justify-center rounded-lg bg-muted">
            <Package className="size-8 text-muted-foreground" />
          </div>

          {/* Product information */}
          <div className="min-w-0 flex-1">
            <div className="flex items-start justify-between gap-3">
              <h3 className="min-w-0 truncate font-medium">
                {product.title}
              </h3>

              <Button
                variant="link"
                render={<a href={`/admin/products/${product.id}`} />}
              >
                View
              </Button>
            </div>

            <p className="mt-2 text-sm font-medium">
              ${product.price}
            </p>

            <div className="mt-2">
              <ProductStatusBadge status={product.status} />
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
