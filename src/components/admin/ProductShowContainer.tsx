"use client"

import { useEffect, useState } from "react"
import { ProductStatusBadge } from "@/components/admin/ProductStatusBadge"

interface Product {
  id: string
  title: string
  slug: string
  description: string
  price: string
  status: "draft" | "active" | "archived"
  createdAt: string
  updatedAt: string
}

interface ProductShowContainerProps {
  productId: string
}

export function ProductShowContainer({
  productId,
}: ProductShowContainerProps) {
  const [product, setProduct] = useState<Product | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadProduct() {
      try {
        const response = await fetch(
          `/api/admin/products/${productId}`
        )

        if (!response.ok) {
          const message = await response.text()
          throw new Error(message || "Failed to load product")
        }

        const data = await response.json()

        setProduct(data)
      } catch (error) {
        console.error("Error loading product:", error)

        setError(
          error instanceof Error
            ? error.message
            : "Failed to load product"
        )
      } finally {
        setIsLoading(false)
      }
    }

    loadProduct()
  }, [productId])

  if (isLoading) {
    return (
      <div className="rounded-xl border bg-card p-10 text-center text-sm text-muted-foreground">
        Loading product...
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-xl border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
        {error}
      </div>
    )
  }

  if (!product) {
    return (
      <div className="rounded-xl border bg-card p-10 text-center text-sm text-muted-foreground">
        Product not found.
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="w-full">
          <h2 className="text-2xl font-semibold tracking-tight">
            {product.title}
          </h2>

          <p className="mt-1 text-sm text-muted-foreground">
            Product details
          </p>
        </div>
        <span className="text-sm font-medium">
          Status:
        </span>
        <ProductStatusBadge status={product.status} />
      </div>

      {/* General */}
      <div className="rounded-xl border bg-card p-6 text-card-foreground shadow-sm">
        <h3 className="text-lg font-semibold">
          General
        </h3>

        <div className="mt-6 space-y-6">
          <div>
            <p className="text-sm text-muted-foreground">
              Title
            </p>
            <p className="mt-1 font-medium">
              {product.title}
            </p>
          </div>

          <div>
            <p className="text-sm text-muted-foreground">
              Slug
            </p>
            <p className="mt-1">
              {product.slug}
            </p>
          </div>

          <div>
            <p className="text-sm text-muted-foreground">
              Description
            </p>
            <p className="mt-1 whitespace-pre-wrap">
              {product.description}
            </p>
          </div>

          <div>
            <p className="text-sm text-muted-foreground">
              Price
            </p>
            <p className="mt-1 font-medium">
              ${product.price}
            </p>
          </div>
        </div>
      </div>

      {/* Information */}
      <div className="rounded-xl border bg-card p-6 text-card-foreground shadow-sm">
        <h3 className="text-lg font-semibold">
          Information
        </h3>

        <div className="mt-6 space-y-6">
          <div>
            <p className="text-sm text-muted-foreground">
              ID
            </p>
            <p className="mt-1 break-all font-mono text-sm">
              {product.id}
            </p>
          </div>

          <div className="grid gap-6 sm:grid-cols-2">
            <div>
              <p className="text-sm text-muted-foreground">
                Created
              </p>
              <p className="mt-1 text-sm">
                {new Date(product.createdAt).toLocaleString()}
              </p>
            </div>

            <div>
              <p className="text-sm text-muted-foreground">
                Updated
              </p>
              <p className="mt-1 text-sm">
                {new Date(product.updatedAt).toLocaleString()}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}