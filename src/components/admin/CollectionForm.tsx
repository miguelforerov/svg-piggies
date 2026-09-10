import { useState } from "react"
import { Button } from "@/components/ui/button"
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
  onCancel: () => void
  initialValues?: CollectionFormData
}

export function CollectionForm({
  onSubmit,
  onCancel,
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

      <div className="flex items-center justify-end gap-3 border-t pt-6">
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit">Save Collection</Button>
      </div>
    </form>
  )
}
