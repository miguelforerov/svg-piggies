"use client"

import { useState } from "react"
import type { Collection } from "@/types/collection"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"

export interface CollectionFormData {
  name: string
  slug: string
  description: string
}

interface CollectionFormProps {
  onSubmit: (data: CollectionFormData) => void
  initialValues?: CollectionFormData
}

function CollectionForm({
  onSubmit,
  initialValues,
}: CollectionFormProps) {
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
          placeholder="Collection name"
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
          placeholder="collection-slug"
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
          placeholder="Collection description"
          rows={6}
        />
      </div>

    </form>
  )
}

interface CollectionFormContainerProps {
  initialValues?: CollectionFormData
  collectionId?: string
  onSuccess?: (collection: Collection) => void
}

export function CollectionFormContainer({
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
