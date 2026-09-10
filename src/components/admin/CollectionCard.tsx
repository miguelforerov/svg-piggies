import { Folder } from "lucide-react"
import type { Collection } from "@/types/collection"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

interface CollectionCardProps {
  collection: Collection
}

export function CollectionCard({ collection }: CollectionCardProps) {
  return (
    <Card className="shadow-sm">
      <CardContent>
        <div className="flex gap-4">
          <div className="flex size-24 shrink-0 items-center justify-center rounded-lg bg-muted">
            <Folder className="size-8 text-muted-foreground" />
          </div>

          <div className="min-w-0 flex-1">
            <div className="flex items-start justify-between gap-3">
              <h3 className="min-w-0 truncate font-medium">
                {collection.name}
              </h3>

              <Button
                variant="link"
                render={<a href={`/admin/collections/${collection.id}`} />}
              >
                View
              </Button>
            </div>

            <p className="mt-2 text-sm text-muted-foreground">
              /{collection.slug}
            </p>
            {collection.description && (
              <p className="mt-2 line-clamp-2 text-sm">
                {collection.description}
              </p>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
