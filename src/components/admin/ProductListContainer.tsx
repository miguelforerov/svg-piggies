"use client"

import { useEffect, useState } from "react"
import { ProductCard } from "@/components/admin/ProductCard"
interface Product {
  id: string
  title: string
  price: string
  status: "draft" | "active" | "archived"
}

export function ProductListContainer() {
  const [products, setProducts] = useState<Product[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadProducts() {
      try {
        const response = await fetch("/api/admin/products")

        if (!response.ok) {
          const message = await response.text()
          throw new Error(message || "Failed to load products")
        }

        const data = await response.json()

        setProducts(data)
      } catch (error) {
        console.error("Error loading products:", error)

        setError(
          error instanceof Error
            ? error.message
            : "Failed to load products"
        )
      } finally {
        setIsLoading(false)
      }
    }

    loadProducts()
  }, [])

  if (isLoading) {
    return (
      <div className="rounded-xl border bg-card p-10 text-center text-sm text-muted-foreground">
        Loading products...
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

  return (
    <div className="space-y-4">
      {products.map((product) => (
        <ProductCard
          key={product.id}
          product={product}
        />
      ))}
    </div>
  )
}