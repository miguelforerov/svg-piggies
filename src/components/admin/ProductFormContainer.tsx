"use client"

import { useEffect, useState } from "react"

import {
  ProductForm,
  type ProductFormData,
  type ProductFormInitialValues,
} from "@/components/admin/ProductForm"
import type { ProductType } from "@/types/product-type"

export interface Product extends ProductFormInitialValues {
  id: string
  createdAt: string
  updatedAt: string
}

interface ProductTypeAssignment {
  productId: string
  productTypeId: string
}

interface ProductFormContainerProps {
  onCancel: () => void
  initialValues?: ProductFormInitialValues
  productId?: string
  onSuccess?: (product: Product) => void
  formId?: string
}

export function ProductFormContainer({
  onCancel,
  initialValues,
  productId,
  onSuccess,
  formId,
}: ProductFormContainerProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isLoadingProductTypes, setIsLoadingProductTypes] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [productTypesLoadError, setProductTypesLoadError] = useState<
    string | null
  >(null)
  const [productTypes, setProductTypes] = useState<ProductType[]>([])
  const [assignedProductTypeIds, setAssignedProductTypeIds] = useState<
    string[]
  >([])
  const isEditing = productId !== undefined

  useEffect(() => {
    let cancelled = false

    async function loadProductTypes() {
      setIsLoadingProductTypes(true)
      setProductTypesLoadError(null)

      try {
        const requests: [Promise<Response>, Promise<Response>?] = [
          fetch("/api/admin/product-types"),
        ]

        if (productId) {
          requests.push(
            fetch(`/api/admin/products/${productId}/product-types`)
          )
        }

        const [productTypesResponse, assignmentsResponse] = await Promise.all(
          requests
        )

        if (!productTypesResponse.ok) {
          const message = await productTypesResponse.text()
          throw new Error(message || "Failed to load product types")
        }

        if (assignmentsResponse && !assignmentsResponse.ok) {
          const message = await assignmentsResponse.text()
          throw new Error(
            message || "Failed to load the product's product types"
          )
        }

        const loadedProductTypes: ProductType[] =
          await productTypesResponse.json()
        const assignments: ProductTypeAssignment[] = assignmentsResponse
          ? await assignmentsResponse.json()
          : []

        if (!cancelled) {
          setProductTypes(loadedProductTypes)
          setAssignedProductTypeIds(
            assignments.map((assignment) => assignment.productTypeId)
          )
        }
      } catch (error) {
        console.error("Error loading product types:", error)

        if (!cancelled) {
          setProductTypesLoadError(
            error instanceof Error
              ? error.message
              : "Failed to load product types"
          )
        }
      } finally {
        if (!cancelled) {
          setIsLoadingProductTypes(false)
        }
      }
    }

    loadProductTypes()

    return () => {
      cancelled = true
    }
  }, [productId])

  async function syncProductTypes(
    targetProductId: string,
    selectedProductTypeIds: string[]
  ) {
    const assignmentsResponse = await fetch(
      `/api/admin/products/${targetProductId}/product-types`
    )

    if (!assignmentsResponse.ok) {
      const message = await assignmentsResponse.text()
      throw new Error(
        message || "Failed to load the product's product types"
      )
    }

    const currentAssignments: ProductTypeAssignment[] =
      await assignmentsResponse.json()
    const currentProductTypeIds = currentAssignments.map(
      (assignment) => assignment.productTypeId
    )
    const assignedIds = new Set(currentProductTypeIds)
    const selectedIds = new Set(selectedProductTypeIds)
    const requests = [
      ...selectedProductTypeIds
        .filter((id) => !assignedIds.has(id))
        .map((productTypeId) =>
          fetch(`/api/admin/products/${targetProductId}/product-types`, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({ productTypeId }),
          })
        ),
      ...currentProductTypeIds
        .filter((id) => !selectedIds.has(id))
        .map((productTypeId) =>
          fetch(
            `/api/admin/products/${targetProductId}/product-types/${productTypeId}`,
            { method: "DELETE" }
          )
        ),
    ]

    const responses = await Promise.all(requests)
    const failedResponse = responses.find((response) => !response.ok)

    if (failedResponse) {
      const message = await failedResponse.text()
      throw new Error(message || "Failed to save product types")
    }

    setAssignedProductTypeIds(selectedProductTypeIds)
  }

  async function handleSubmit(data: ProductFormData) {
    setIsSubmitting(true)
    setError(null)

    try {
      const { productTypeIds, ...productData } = data
      const response = await fetch(
        isEditing
          ? `/api/admin/products/${productId}`
          : "/api/admin/products",
        {
          method: isEditing ? "PUT" : "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(productData),
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

      const savedProduct: Product = await response.json()

      await syncProductTypes(savedProduct.id, productTypeIds)

      if (isEditing) {
        const updatedProduct = savedProduct
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
      {(error || productTypesLoadError) && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
          {error || productTypesLoadError}
        </div>
      )}

      {isLoadingProductTypes ? (
        <p className="text-sm text-muted-foreground">
          Loading product types...
        </p>
      ) : !productTypesLoadError ? (
        <ProductForm
          onSubmit={handleSubmit}
          onCancel={onCancel}
          initialValues={initialValues}
          productTypes={productTypes}
          initialProductTypeIds={assignedProductTypeIds}
          isSubmitting={isSubmitting}
          formId={formId}
        />
      ) : null}
    </div>
  )
}
