import { Tags } from "lucide-react"
import type { ProductType } from "@/types/product-type"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

interface ProductTypeCardProps {
  productType: ProductType
}

export function ProductTypeCard({ productType }: ProductTypeCardProps) {
  return (
    <Card className="shadow-sm">
      <CardContent>
        <div className="flex gap-4">
          <div className="flex size-24 shrink-0 items-center justify-center rounded-lg bg-muted">
            <Tags className="size-8 text-muted-foreground" />
          </div>

          <div className="min-w-0 flex-1">
            <div className="flex items-start justify-between gap-3">
              <h3 className="min-w-0 truncate font-medium">
                {productType.name}
              </h3>

              <Button
                variant="link"
                render={
                  <a href={`/admin/product-types/${productType.id}`} />
                }
              >
                View
              </Button>
            </div>

            <p className="mt-2 text-sm text-muted-foreground">
              /{productType.slug}
            </p>
            {productType.description && (
              <p className="mt-2 line-clamp-2 text-sm">
                {productType.description}
              </p>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
