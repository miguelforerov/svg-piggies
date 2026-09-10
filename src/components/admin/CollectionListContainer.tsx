"use client"

import { useEffect, useState } from "react"
import type { Collection } from "@/types/collection"
import { CollectionCard } from "@/components/admin/CollectionCard"
import { Card, CardContent } from "@/components/ui/card"

export function CollectionListContainer() {
  const [collections, setCollections] = useState<Collection[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadCollections() {
      try {
        const response = await fetch("/api/admin/collections")

        if (!response.ok) {
          const message = await response.text()
          throw new Error(message || "Failed to load collections")
        }

        const data: Collection[] = await response.json()
        setCollections(data)
      } catch (error) {
        console.error("Error loading collections:", error)
        setError(
          error instanceof Error
            ? error.message
            : "Failed to load collections"
        )
      } finally {
        setIsLoading(false)
      }
    }

    loadCollections()
  }, [])

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-10 text-center text-muted-foreground">
          Loading collections...
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <div className="rounded-xl border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
        {error}
      </div>
    )
  }

  if (collections.length === 0) {
    return (
      <Card>
        <CardContent className="p-10 text-center text-muted-foreground">
          No collections found.
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      {collections.map((collection) => (
        <CollectionCard key={collection.id} collection={collection} />
      ))}
    </div>
  )
}
