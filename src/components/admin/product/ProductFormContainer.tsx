"use client"

import { useEffect, useState } from "react"

import {
  ProductForm,
  type ProductFormData,
  type ProductFormInitialValues,
} from "@/components/admin/product/ProductForm"
import type { ProductType } from "@/types/product-type"
import type { Collection } from "@/types/collection"

export interface Product extends ProductFormInitialValues {
  id: string
  createdAt: string
  updatedAt: string
}

interface ProductTypeAssignment {
  productId: string
  productTypeId: string
}

interface CollectionAssignment {
  productId: string
  collectionId: string
}

interface ProductFormContainerProps {
  initialValues?: ProductFormInitialValues
  productId?: string
  onSuccess?: (product: Product) => void
  formId?: string
}

export function ProductFormContainer({
  initialValues,
  productId,
  onSuccess,
  formId,
}: ProductFormContainerProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isLoadingOptions, setIsLoadingOptions] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [optionsLoadError, setOptionsLoadError] = useState<
    string | null
  >(null)
  const [productTypes, setProductTypes] = useState<ProductType[]>([])
  const [collections, setCollections] = useState<Collection[]>([])
  const [assignedProductTypeIds, setAssignedProductTypeIds] = useState<
    string[]
  >([])
  const [assignedCollectionIds, setAssignedCollectionIds] = useState<
    string[]
  >([])
  const isEditing = productId !== undefined

  useEffect(() => {
    let cancelled = false

    async function loadOptions() {
      setIsLoadingOptions(true)
      setOptionsLoadError(null)

      try {
        const requests = [
          fetch("/api/admin/product-types"),
          fetch("/api/admin/collections"),
          productId
            ? fetch(`/api/admin/products/${productId}/product-types`)
            : Promise.resolve(null),
          productId
            ? fetch(`/api/admin/products/${productId}/collections`)
            : Promise.resolve(null),
        ] as const

        const [
          productTypesResponse,
          collectionsResponse,
          productTypeAssignmentsResponse,
          collectionAssignmentsResponse,
        ] = await Promise.all(requests)

        if (!productTypesResponse.ok) {
          const message = await productTypesResponse.text()
          throw new Error(message || "Failed to load product types")
        }

        if (!collectionsResponse.ok) {
          const message = await collectionsResponse.text()
          throw new Error(message || "Failed to load collections")
        }

        if (
          productTypeAssignmentsResponse &&
          !productTypeAssignmentsResponse.ok
        ) {
          const message = await productTypeAssignmentsResponse.text()
          throw new Error(
            message || "Failed to load the product's product types"
          )
        }

        if (
          collectionAssignmentsResponse &&
          !collectionAssignmentsResponse.ok
        ) {
          const message = await collectionAssignmentsResponse.text()
          throw new Error(
            message || "Failed to load the product's collections"
          )
        }

        const loadedProductTypes: ProductType[] =
          await productTypesResponse.json()
        const loadedCollections: Collection[] =
          await collectionsResponse.json()
        const productTypeAssignments: ProductTypeAssignment[] =
          productTypeAssignmentsResponse
            ? await productTypeAssignmentsResponse.json()
            : []
        const collectionAssignments: CollectionAssignment[] =
          collectionAssignmentsResponse
            ? await collectionAssignmentsResponse.json()
            : []

        if (!cancelled) {
          setProductTypes(loadedProductTypes)
          setCollections(loadedCollections)
          setAssignedProductTypeIds(
            productTypeAssignments.map(
              (assignment) => assignment.productTypeId
            )
          )
          setAssignedCollectionIds(
            collectionAssignments.map(
              (assignment) => assignment.collectionId
            )
          )
        }
      } catch (error) {
        console.error("Error loading product options:", error)

        if (!cancelled) {
          setOptionsLoadError(
            error instanceof Error
              ? error.message
              : "Failed to load product options"
          )
        }
      } finally {
        if (!cancelled) {
          setIsLoadingOptions(false)
        }
      }
    }

    loadOptions()

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

  async function syncCollections(
    targetProductId: string,
    selectedCollectionIds: string[]
  ) {
    const assignmentsResponse = await fetch(
      `/api/admin/products/${targetProductId}/collections`
    )

    if (!assignmentsResponse.ok) {
      const message = await assignmentsResponse.text()
      throw new Error(
        message || "Failed to load the product's collections"
      )
    }

    const currentAssignments: CollectionAssignment[] =
      await assignmentsResponse.json()
    const currentCollectionIds = currentAssignments.map(
      (assignment) => assignment.collectionId
    )
    const assignedIds = new Set(currentCollectionIds)
    const selectedIds = new Set(selectedCollectionIds)
    const requests = [
      ...selectedCollectionIds
        .filter((id) => !assignedIds.has(id))
        .map((collectionId) =>
          fetch(`/api/admin/products/${targetProductId}/collections`, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({ collectionId }),
          })
        ),
      ...currentCollectionIds
        .filter((id) => !selectedIds.has(id))
        .map((collectionId) =>
          fetch(
            `/api/admin/products/${targetProductId}/collections/${collectionId}`,
            { method: "DELETE" }
          )
        ),
    ]

    const responses = await Promise.all(requests)
    const failedResponse = responses.find((response) => !response.ok)

    if (failedResponse) {
      const message = await failedResponse.text()
      throw new Error(message || "Failed to save collections")
    }

    setAssignedCollectionIds(selectedCollectionIds)
  }

  async function handleSubmit(data: ProductFormData) {
    if (isSubmitting) return

    setIsSubmitting(true)
    setError(null)

    try {
      const { productTypeIds, collectionIds, ...productData } = data
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

      await Promise.all([
        syncProductTypes(savedProduct.id, productTypeIds),
        syncCollections(savedProduct.id, collectionIds),
      ])

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
      {(error || optionsLoadError) && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
          {error || optionsLoadError}
        </div>
      )}

      {isLoadingOptions ? (
        <p className="text-sm text-muted-foreground">
          Loading product options...
        </p>
      ) : !optionsLoadError ? (
        <ProductForm
          onSubmit={handleSubmit}
          initialValues={initialValues}
          productTypes={productTypes}
          initialProductTypeIds={assignedProductTypeIds}
          collections={collections}
          initialCollectionIds={assignedCollectionIds}
          formId={formId}
        />
      ) : null}
    </div>
  )
}
