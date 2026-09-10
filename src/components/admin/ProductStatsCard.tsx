"use client"

import { useEffect, useState } from "react"
import { Package } from "lucide-react"
import { StatsCard } from "@/components/admin/StatsCard"

export function ProductStatsCard() {
  const [totalProducts, setTotalProducts] = useState<number | null>(null)
  const [error, setError] = useState(false)

  useEffect(() => {
    let isCancelled = false

    async function loadTotalProducts() {
      try {
        const response = await fetch("/api/admin/products")

        if (!response.ok) {
          throw new Error("Failed to load products")
        }

        const products: unknown = await response.json()

        if (!Array.isArray(products)) {
          throw new Error("Invalid products response")
        }

        if (!isCancelled) {
          setTotalProducts(products.length)
        }
      } catch (error) {
        console.error("Error loading product statistics:", error)

        if (!isCancelled) {
          setError(true)
        }
      }
    }

    loadTotalProducts()

    return () => {
      isCancelled = true
    }
  }, [])

  return (
    <StatsCard
      title="Products"
      value={error ? "—" : (totalProducts ?? "...")}
      icon={Package}
      description={error ? "Unable to load products" : "Total products"}
    />
  )
}
