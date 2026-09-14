"use client"

import { useState } from "react"
import type { ProductType } from "@/types/product-type"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"

export interface ProductTypeFormData {
  name: string
  slug: string
  description: string
}

interface ProductTypeFormProps {
  onSubmit: (data: ProductTypeFormData) => void
  initialValues?: ProductTypeFormData
}

function ProductTypeForm({
  onSubmit,
  initialValues,
}: ProductTypeFormProps) {
  const [name, setName] = useState(initialValues?.name ?? "")
  const [slug, setSlug] = useState(initialValues?.slug ?? "")
  const [description, setDescription] = useState(
    initialValues?.description ?? ""
  )

  function handleSubmit(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault()

    onSubmit({
      name,
      slug,
      description,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="name">Name</Label>
        <Input
          id="name"
          name="name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="Product type name"
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="slug">Slug</Label>
        <Input
          id="slug"
          name="slug"
          value={slug}
          onChange={(event) => setSlug(event.target.value)}
          placeholder="product-type-slug"
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          name="description"
          value={description}
          onChange={(event) => setDescription(event.target.value)}
          placeholder="Product type description"
          rows={6}
        />
      </div>
    </form>
  )
}

interface ProductTypeFormContainerProps {
  onCancel: () => void
  initialValues?: ProductTypeFormData
  productTypeId?: string
  onSuccess?: (productType: ProductType) => void
}

export function ProductTypeFormContainer({
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
