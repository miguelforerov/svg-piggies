"use client"

import { useEffect, useState } from "react"
import {
  Archive,
  EllipsisVertical,
  FilePenLine,
  Pencil,
  Send,
  Trash2,
} from "lucide-react"
import {
  ProductFormContainer,
  type Product,
} from "@/components/admin/ProductFormContainer"
import { ProductStatusBadge } from "@/components/admin/ProductStatusBadge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface ProductShowContainerProps {
  productId: string
}

export function ProductShowContainer({
  productId,
}: ProductShowContainerProps) {
  const [product, setProduct] = useState<Product | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<
    "publish" | "draft" | "archive" | "delete" | null
  >(null)
  const [actionError, setActionError] = useState<string | null>(null)

  useEffect(() => {
    async function loadProduct() {
      try {
        const response = await fetch(
          `/api/admin/products/${productId}`
        )

        if (!response.ok) {
          const message = await response.text()
          throw new Error(message || "Failed to load product")
        }

        const data: Product = await response.json()

        setProduct(data)
      } catch (error) {
        console.error("Error loading product:", error)

        setError(
          error instanceof Error
            ? error.message
            : "Failed to load product"
        )
      } finally {
        setIsLoading(false)
      }
    }

    loadProduct()
  }, [productId])

  async function updateStatus(status: Product["status"]) {
    if (!product) {
      return
    }

    setPendingAction(
      status === "active"
        ? "publish"
        : status === "draft"
          ? "draft"
          : "archive"
    )
    setActionError(null)

    try {
      const response = await fetch(`/api/admin/products/${productId}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          title: product.title,
          slug: product.slug,
          description: product.description,
          price: product.price,
          status,
        }),
      })

      if (!response.ok) {
        const message = await response.text()
        throw new Error(message || "Failed to update product status")
      }

      const updatedProduct: Product = await response.json()
      setProduct(updatedProduct)
    } catch (error) {
      console.error("Error updating product status:", error)
      setActionError(
        error instanceof Error
          ? error.message
          : "Failed to update product status"
      )
    } finally {
      setPendingAction(null)
    }
  }

  async function deleteProduct() {
    setPendingAction("delete")
    setActionError(null)

    try {
      const response = await fetch(`/api/admin/products/${productId}`, {
        method: "DELETE",
      })

      if (!response.ok) {
        const message = await response.text()
        throw new Error(message || "Failed to delete product")
      }

      window.location.href = "/admin/products"
    } catch (error) {
      console.error("Error deleting product:", error)
      setActionError(
        error instanceof Error
          ? error.message
          : "Failed to delete product"
      )
      setPendingAction(null)
    }
  }

  if (isLoading) {
    return (
      <div className="rounded-xl border bg-card p-10 text-center text-sm text-muted-foreground">
        Loading product...
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-xl border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
        {error}
      </div>
    )
  }

  if (!product) {
    return (
      <div className="rounded-xl border bg-card p-10 text-center text-sm text-muted-foreground">
        Product not found.
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="w-full">
          <h2 className="text-2xl font-semibold tracking-tight">
            {product.title}
          </h2>

          <p className="mt-1 text-sm text-muted-foreground">
            Product details
          </p>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm font-medium">
            Status:
          </span>
          <ProductStatusBadge status={product.status} />
        </div>
      </div>

      <Dialog
        open={isEditing}
        onOpenChange={setIsEditing}
        disablePointerDismissal
      >
        <DialogContent
          className="max-h-[90vh] overflow-y-auto sm:max-w-3xl"
          showCloseButton={false}
        >
          <DialogHeader>
            <DialogTitle>Edit Product</DialogTitle>
          </DialogHeader>

          <ProductFormContainer
            productId={productId}
            initialValues={product}
            onCancel={() => setIsEditing(false)}
            onSuccess={(updatedProduct) => {
              setProduct(updatedProduct)
              setIsEditing(false)
            }}
          />
        </DialogContent>
      </Dialog>

      <Dialog
        open={isDeleteDialogOpen}
        onOpenChange={(open) => {
          if (pendingAction !== "delete") {
            setIsDeleteDialogOpen(open)
            setActionError(null)
          }
        }}
        disablePointerDismissal
      >
        <DialogContent showCloseButton={false}>
          <DialogHeader>
            <DialogTitle>Delete Product</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete “{product.title}”? This action
              cannot be undone.
            </DialogDescription>
          </DialogHeader>

          {actionError && (
            <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
              {actionError}
            </div>
          )}

          <div className="flex items-center justify-end gap-3 border-t pt-4">
            <Button
              variant="outline"
              disabled={pendingAction === "delete"}
              onClick={() => {
                setIsDeleteDialogOpen(false)
                setActionError(null)
              }}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              disabled={pendingAction === "delete"}
              onClick={deleteProduct}
            >
              <Trash2 />
              {pendingAction === "delete" ? "Deleting..." : "Delete"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {actionError && !isDeleteDialogOpen && (
        <div className="rounded-xl border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
          {actionError}
        </div>
      )}

      {/* General */}
      <div className="rounded-xl border bg-card p-6 text-card-foreground shadow-sm">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-lg font-semibold">
            General
          </h3>

          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <Button
                  variant="secondary"
                  size="icon"
                  aria-label="Product actions"
                />
              }
            >
              <EllipsisVertical />
            </DropdownMenuTrigger>

            <DropdownMenuContent>
              {product.status !== "archived" && (
                <DropdownMenuItem onClick={() => setIsEditing(true)}>
                  <Pencil />
                  Edit
                </DropdownMenuItem>
              )}
              <DropdownMenuItem
                disabled={pendingAction !== null}
                onClick={() =>
                  updateStatus(
                    product.status === "draft" ? "active" : "draft"
                  )
                }
              >
                {product.status !== "draft" ? <FilePenLine /> : <Send />}
                {product.status !== "draft"
                  ? pendingAction === "draft"
                    ? "Moving to draft..."
                    : "Move to draft"
                  : pendingAction === "publish"
                    ? "Publishing..."
                    : "Publish"}
              </DropdownMenuItem>
              {product.status !== "archived" && (
                <DropdownMenuItem
                  disabled={pendingAction !== null}
                  onClick={() => updateStatus("archived")}
                >
                  <Archive />
                  {pendingAction === "archive"
                    ? "Archiving..."
                    : "Archive"}
                </DropdownMenuItem>
              )}
              <DropdownMenuItem
                variant="destructive"
                disabled={pendingAction !== null}
                onClick={() => {
                  setActionError(null)
                  setIsDeleteDialogOpen(true)
                }}
              >
                <Trash2 />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        <div className="mt-6 grid gap-6 lg:grid-cols-2">
          {/* Left column */}
          <div className="space-y-6">
            <div>
              <p className="text-sm text-muted-foreground">
                Title
              </p>
              <p className="mt-1 font-medium">
                {product.title}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">
                Price
              </p>
              <p className="mt-1 font-medium">
                ${product.price}
              </p>
            </div>
          </div>

          {/* Right column */}
          <div>
            <p className="text-sm text-muted-foreground">
              Slug
            </p>
            <p className="mt-1">
              {product.slug}
            </p>
          </div>

        </div>
        <div className="mt-6">
          <p className="text-sm text-muted-foreground">
            Description
          </p>
          <p className="mt-1 whitespace-pre-wrap">
            {product.description}
          </p>
        </div>
      </div>

      {/* Information */}
      <div className="rounded-xl border bg-card p-6 text-card-foreground shadow-sm">
        <h3 className="text-lg font-semibold">
          Information
        </h3>

        <div className="mt-6 space-y-6">
          <div>
            <p className="text-sm text-muted-foreground">
              ID
            </p>
            <p className="mt-1 break-all font-mono text-sm">
              {product.id}
            </p>
          </div>

          <div className="grid gap-6 sm:grid-cols-2">
            <div>
              <p className="text-sm text-muted-foreground">
                Created
              </p>
              <p className="mt-1 text-sm">
                {new Date(product.createdAt).toLocaleString()}
              </p>
            </div>

            <div>
              <p className="text-sm text-muted-foreground">
                Updated
              </p>
              <p className="mt-1 text-sm">
                {new Date(product.updatedAt).toLocaleString()}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
