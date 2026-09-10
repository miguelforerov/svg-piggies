"use client"

import { useState } from "react"
import type { ProductType } from "@/types/product-type"
import {
  ProductTypeForm,
  type ProductTypeFormData,
} from "@/components/admin/ProductTypeForm"

interface ProductTypeFormContainerProps {
  onCancel: () => void
  initialValues?: ProductTypeFormData
  productTypeId?: string
  onSuccess?: (productType: ProductType) => void
}

export function ProductTypeFormContainer({
  onCancel,
  initialValues,
  productTypeId,
  onSuccess,
}: ProductTypeFormContainerProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const isEditing = productTypeId !== undefined

  async function handleSubmit(data: ProductTypeFormData) {
    setIsSubmitting(true)
    setError(null)

    try {
      const response = await fetch(
        isEditing
          ? `/api/admin/product-types/${productTypeId}`
          : "/api/admin/product-types",
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
              ? "Failed to update product type"
              : "Failed to create product type")
        )
      }

      if (isEditing) {
        const updatedProductType: ProductType = await response.json()
        onSuccess?.(updatedProductType)
        return
      }

      window.location.href = "/admin/product-types"
    } catch (error) {
      console.error(
        `Error ${isEditing ? "updating" : "creating"} product type:`,
        error
      )
      setError(
        error instanceof Error
          ? error.message
          : isEditing
            ? "Failed to update product type"
            : "Failed to create product type"
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

      <ProductTypeForm
        onSubmit={handleSubmit}
        onCancel={onCancel}
        initialValues={initialValues}
      />

      {isSubmitting && (
        <p className="text-sm text-muted-foreground">
          {isEditing
            ? "Updating product type..."
            : "Creating product type..."}
        </p>
      )}
    </div>
  )
}
