"use client"

import { useState } from "react"

import {
  ProductForm,
  type ProductFormData,
} from "@/components/admin/ProductForm"

interface ProductFormContainerProps {
  onCancel: () => void
}

export function ProductFormContainer({onCancel,}: ProductFormContainerProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(data: ProductFormData) {
    setIsSubmitting(true)
    setError(null)

    try {
      const response = await fetch("/api/admin/products", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      })

      if (!response.ok) {
        const message = await response.text()
        throw new Error(message || "Failed to create product")
      }

      window.location.href = "/admin/products"
    } catch (error) {
      console.error("Error creating product:", error)

      setError(
        error instanceof Error
          ? error.message
          : "Failed to create product"
      )
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="space-y-4">
      {error && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
          {error}
        </div>
      )}

      <ProductForm
        onSubmit={handleSubmit}
        onCancel={onCancel}
      />

      {isSubmitting && (
        <p className="text-sm text-muted-foreground">
          Creating product...
        </p>
      )}
    </div>
  )
}