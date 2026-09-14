import { useState } from "react"
import {
  PRODUCT_STATUSES,
  type ProductStatus,
} from "@/types/product"
import type { ProductType } from "@/types/product-type"
import type { Collection } from "@/types/collection"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

export interface ProductFormData {
  title: string
  slug: string
  description: string
  price: string
  status: ProductStatus
  productTypeIds: string[]
  collectionIds: string[]
}

export type ProductFormInitialValues = Omit<
  ProductFormData,
  "productTypeIds" | "collectionIds"
>

const productStatusLabels: Record<ProductStatus, string> = {
  draft: "Draft",
  active: "Active",
  archived: "Archived",
}

interface ProductFormProps {
  onSubmit: (data: ProductFormData) => void
  initialValues?: ProductFormInitialValues
  productTypes: ProductType[]
  initialProductTypeIds?: string[]
  collections: Collection[]
  initialCollectionIds?: string[]
  formId?: string
}

export function ProductForm({
  onSubmit,
  initialValues,
  productTypes,
  initialProductTypeIds = [],
  collections,
  initialCollectionIds = [],
  formId,
}: ProductFormProps) {
  const [title, setTitle] = useState(initialValues?.title ?? "")
  const [slug, setSlug] = useState(initialValues?.slug ?? "")
  const [description, setDescription] = useState(
    initialValues?.description ?? ""
  )
  const [price, setPrice] = useState(initialValues?.price ?? "")
  const [status, setStatus] = useState<ProductStatus>(
    initialValues?.status ?? "draft"
  )
  const [productTypeIds, setProductTypeIds] = useState<string[]>(
    initialProductTypeIds
  )
  const [collectionIds, setCollectionIds] = useState<string[]>(
    initialCollectionIds
  )

  function toggleProductType(productTypeId: string) {
    setProductTypeIds((currentIds) =>
      currentIds.includes(productTypeId)
        ? currentIds.filter((id) => id !== productTypeId)
        : [...currentIds, productTypeId]
    )
  }

  function toggleCollection(collectionId: string) {
    setCollectionIds((currentIds) =>
      currentIds.includes(collectionId)
        ? currentIds.filter((id) => id !== collectionId)
        : [...currentIds, collectionId]
    )
  }

  function handleSubmit(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault()

    onSubmit({
      title,
      slug,
      description,
      price,
      status,
      productTypeIds,
      collectionIds,
    })
  }

  return (
    <form id={formId} onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="title">Title</Label>

        <Input
          id="title"
          name="title"
          value={title}
          onChange={(event) => setTitle(event.target.value)}
          placeholder="Product title"
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
          placeholder="product-slug"
          required
        />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="price">Price</Label>

          <Input
            id="price"
            name="price"
            type="text"
            value={price}
            onChange={(event) => setPrice(event.target.value)}
            placeholder="4.99"
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="status">Status</Label>

          <Select
            value={status}
            onValueChange={(value) =>
              setStatus(value as ProductStatus)
            }
          >
            <SelectTrigger id="status">
              <SelectValue placeholder="Select status" />
            </SelectTrigger>

            <SelectContent>
              {PRODUCT_STATUSES.map((productStatus) => (
                <SelectItem key={productStatus} value={productStatus}>
                  {productStatusLabels[productStatus]}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <fieldset className="space-y-3">
        <legend className="text-sm font-medium">Product Types</legend>

        {productTypes.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No product types available.
          </p>
        ) : (
          <div className="grid gap-3 rounded-md border p-4 sm:grid-cols-2 max-h-[96px] overflow-hidden overflow-scroll">
            {productTypes.map((productType) => (
              <label
                key={productType.id}
                className="flex cursor-pointer items-start gap-3 text-sm"
              >
                <input
                  type="checkbox"
                  name="productTypeIds"
                  value={productType.id}
                  checked={productTypeIds.includes(productType.id)}
                  onChange={() => toggleProductType(productType.id)}
                  className="mt-0.5 size-4 rounded border-input accent-primary"
                />
                <span>
                  <span className="block font-medium">{productType.name}</span>
                </span>
              </label>
            ))}
          </div>
        )}
      </fieldset>

      <fieldset className="space-y-3">
        <legend className="text-sm font-medium">Collections</legend>

        {collections.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No collections available.
          </p>
        ) : (
          <div className="grid max-h-[96px] gap-3 overflow-scroll rounded-md border p-4 sm:grid-cols-2">
            {collections.map((collection) => (
              <label
                key={collection.id}
                className="flex cursor-pointer items-start gap-3 text-sm"
              >
                <input
                  type="checkbox"
                  name="collectionIds"
                  value={collection.id}
                  checked={collectionIds.includes(collection.id)}
                  onChange={() => toggleCollection(collection.id)}
                  className="mt-0.5 size-4 rounded border-input accent-primary"
                />
                <span className="block font-medium">{collection.name}</span>
              </label>
            ))}
          </div>
        )}
      </fieldset>

      <div className="space-y-2">
        <Label htmlFor="description">Description</Label>

        <Textarea
          id="description"
          name="description"
          value={description}
          onChange={(event) => setDescription(event.target.value)}
          placeholder="Product description"
          rows={6}
        />
      </div>


    </form>
  )
}
