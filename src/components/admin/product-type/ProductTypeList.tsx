"use client"

import { useEffect, useState } from "react"
import type { ProductType } from "@/types/product-type"
import { ProductTypeCard } from "@/components/admin/product-type/ProductTypeCard"
import { Card, CardContent } from "@/components/ui/card"

export function ProductTypeList() {
  const [productTypes, setProductTypes] = useState<ProductType[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadProductTypes() {
      try {
        const response = await fetch("/api/admin/product-types")

        if (!response.ok) {
          const message = await response.text()
          throw new Error(message || "Failed to load product types")
        }

        const data: ProductType[] = await response.json()
        setProductTypes(data)
      } catch (error) {
        console.error("Error loading product types:", error)
        setError(
          error instanceof Error
            ? error.message
            : "Failed to load product types"
        )
      } finally {
        setIsLoading(false)
      }
    }

    loadProductTypes()
  }, [])

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-10 text-center text-muted-foreground">
          Loading product types...
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

  if (productTypes.length === 0) {
    return (
      <Card>
        <CardContent className="p-10 text-center text-muted-foreground">
          No product types found.
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      {productTypes.map((productType) => (
        <ProductTypeCard key={productType.id} productType={productType} />
      ))}
    </div>
  )
}
