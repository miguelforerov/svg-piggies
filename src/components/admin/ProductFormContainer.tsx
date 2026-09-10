"use client"

import { useState } from "react"

import {
  ProductForm,
  type ProductFormData,
} from "@/components/admin/ProductForm"

export interface Product extends ProductFormData {
  id: string
  createdAt: string
  updatedAt: string
}

interface ProductFormContainerProps {
  onCancel: () => void
  initialValues?: ProductFormData
  productId?: string
  onSuccess?: (product: Product) => void
}

export function ProductFormContainer({
  onCancel,
  initialValues,
  productId,
  onSuccess,
}: ProductFormContainerProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const isEditing = productId !== undefined

  async function handleSubmit(data: ProductFormData) {
    setIsSubmitting(true)
    setError(null)

    try {
      const response = await fetch(
        isEditing
          ? `/api/admin/products/${productId}`
          : "/api/admin/products",
        {
          method: isEditing ? "PUT" : "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(data),
        }
      )

      if (!response.ok) {
        const message = await response.text()
        throw new Error(
          message ||
            (isEditing
              ? "Failed to update product"
              : "Failed to create product")
        )
      }

      if (isEditing) {
        const updatedProduct: Product = await response.json()
        onSuccess?.(updatedProduct)
        return
      }

      window.location.href = "/admin/products"
    } catch (error) {
      console.error(
        `Error ${isEditing ? "updating" : "creating"} product:`,
        error
      )

      setError(
        error instanceof Error
          ? error.message
          : isEditing
            ? "Failed to update product"
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
        initialValues={initialValues}
      />

      {isSubmitting && (
        <p className="text-sm text-muted-foreground">
          {isEditing ? "Updating product..." : "Creating product..."}
        </p>
      )}
    </div>
  )
}
