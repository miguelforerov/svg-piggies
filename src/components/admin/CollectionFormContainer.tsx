"use client"

import { useState } from "react"
import type { Collection } from "@/types/collection"
import {
  CollectionForm,
  type CollectionFormData,
} from "@/components/admin/CollectionForm"

interface CollectionFormContainerProps {
  onCancel: () => void
  initialValues?: CollectionFormData
  collectionId?: string
  onSuccess?: (collection: Collection) => void
}

export function CollectionFormContainer({
  onCancel,
  initialValues,
  collectionId,
  onSuccess,
}: CollectionFormContainerProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const isEditing = collectionId !== undefined

  async function handleSubmit(data: CollectionFormData) {
    setIsSubmitting(true)
    setError(null)

    try {
      const response = await fetch(
        isEditing
          ? `/api/admin/collections/${collectionId}`
          : "/api/admin/collections",
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
              ? "Failed to update collection"
              : "Failed to create collection")
        )
      }

      if (isEditing) {
        const updatedCollection: Collection = await response.json()
        onSuccess?.(updatedCollection)
        return
      }

      window.location.href = "/admin/collections"
    } catch (error) {
      console.error(
        `Error ${isEditing ? "updating" : "creating"} collection:`,
        error
      )
      setError(
        error instanceof Error
          ? error.message
          : isEditing
            ? "Failed to update collection"
            : "Failed to create collection"
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

      <CollectionForm
        onSubmit={handleSubmit}
        onCancel={onCancel}
        initialValues={initialValues}
      />

      {isSubmitting && (
        <p className="text-sm text-muted-foreground">
          {isEditing ? "Updating collection..." : "Creating collection..."}
        </p>
      )}
    </div>
  )
}
