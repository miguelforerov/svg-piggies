"use client"

import { useEffect, useState } from "react"
import { Folder, Package, Tags, type LucideIcon } from "lucide-react"
import { StatsCard } from "@/components/admin/StatsCard"

type ResourceKey = "products" | "collections" | "productTypes"

interface ResourceConfig {
  key: ResourceKey
  endpoint: string
  title: string
  description: string
  errorDescription: string
  icon: LucideIcon
}

const resources: ResourceConfig[] = [
  {
    key: "products",
    endpoint: "/api/admin/products",
    title: "Products",
    description: "Total products",
    errorDescription: "Unable to load products",
    icon: Package,
  },
  {
    key: "collections",
    endpoint: "/api/admin/collections",
    title: "Collections",
    description: "Total collections",
    errorDescription: "Unable to load collections",
    icon: Folder,
  },
  {
    key: "productTypes",
    endpoint: "/api/admin/product-types",
    title: "Product Types",
    description: "Total product types",
    errorDescription: "Unable to load product types",
    icon: Tags,
  },
]

export function DashboardStats() {
  const [totals, setTotals] = useState<Partial<Record<ResourceKey, number>>>(
    {}
  )
  const [errors, setErrors] = useState<Partial<Record<ResourceKey, boolean>>>(
    {}
  )

  useEffect(() => {
    let isCancelled = false

    async function loadTotal(resource: ResourceConfig) {
      try {
        const response = await fetch(resource.endpoint)

        if (!response.ok) {
          throw new Error(`Failed to load ${resource.title.toLowerCase()}`)
        }

        const data: unknown = await response.json()

        if (!Array.isArray(data)) {
          throw new Error(
            `Invalid ${resource.title.toLowerCase()} response`
          )
        }

        if (!isCancelled) {
          setTotals((currentTotals) => ({
            ...currentTotals,
            [resource.key]: data.length,
          }))
        }
      } catch (error) {
        console.error(`Error loading ${resource.key} statistics:`, error)

        if (!isCancelled) {
          setErrors((currentErrors) => ({
            ...currentErrors,
            [resource.key]: true,
          }))
        }
      }
    }

    resources.forEach((resource) => {
      loadTotal(resource)
    })

    return () => {
      isCancelled = true
    }
  }, [])

  return (
    <section className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {resources.map((resource) => {
        const hasError = errors[resource.key] === true

        return (
          <StatsCard
            key={resource.key}
            title={resource.title}
            value={hasError ? "—" : (totals[resource.key] ?? "...")}
            icon={resource.icon}
            description={
              hasError ? resource.errorDescription : resource.description
            }
          />
        )
      })}
    </section>
  )
}

