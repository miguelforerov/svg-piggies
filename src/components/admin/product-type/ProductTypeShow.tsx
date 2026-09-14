"use client"

import { useEffect, useState } from "react"
import { EllipsisVertical, Pencil, Trash2Icon } from "lucide-react"
import type { ProductType } from "@/types/product-type"
import { DELETE_CONFIRMATION_TEXT } from "@/lib/constants"
import { ProductTypeFormContainer } from "@/components/admin/product-type/ProductTypeFormContainer"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogCancel,
  AlertDialogTrigger,
  AlertDialogFooter,
  AlertDialogAction,
  AlertDialogMedia,
} from "@/components/ui/alert-dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface ProductTypeShowContainerProps {
  productTypeId: string
}

export function ProductTypeShow({
  productTypeId,
}: ProductTypeShowContainerProps) {
  const [productType, setProductType] = useState<ProductType | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  useEffect(() => {
    async function loadProductType() {
      try {
        const response = await fetch(
          `/api/admin/product-types/${productTypeId}`
        )

        if (!response.ok) {
          const message = await response.text()
          throw new Error(message || "Failed to load product type")
        }

        const data: ProductType = await response.json()
        setProductType(data)
      } catch (error) {
        console.error("Error loading product type:", error)
        setError(
          error instanceof Error
            ? error.message
            : "Failed to load product type"
        )
      } finally {
        setIsLoading(false)
      }
    }

    loadProductType()
  }, [productTypeId])

  async function deleteProductType() {
    setIsDeleting(true)
    setDeleteError(null)

    try {
      const response = await fetch(
        `/api/admin/product-types/${productTypeId}`,
        { method: "DELETE" }
      )

      if (!response.ok) {
        const message = await response.text()
        throw new Error(message || "Failed to delete product type")
      }

      window.location.href = "/admin/product-types"
    } catch (error) {
      console.error("Error deleting product type:", error)
      setDeleteError(
        error instanceof Error
          ? error.message
          : "Failed to delete product type"
      )
      setIsDeleting(false)
    }
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-10 text-center text-muted-foreground">
          Loading product type...
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <div className="rounded-xl border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
        {error}
      </div>
    )
  }

  if (!productType) {
    return (
      <Card>
        <CardContent className="p-10 text-center text-muted-foreground">
          Product type not found.
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold tracking-tight">
          {productType.name}
        </h2>
      </div>

      <Dialog
        open={isEditing}
        onOpenChange={setIsEditing}
        disablePointerDismissal
      >
        <DialogContent
          className="max-h-[90vh] overflow-y-auto sm:max-w-3xl"
          showDialogFooter={true}
          onCancel={() => {
            setIsDeleteDialogOpen(false)
            setDeleteError(null)
          }}
          // onSave={onSave}
        >
          <DialogHeader>
            <DialogTitle>Edit Product Type</DialogTitle>
          </DialogHeader>

          <ProductTypeFormContainer
            productTypeId={productTypeId}
            initialValues={productType}
            onCancel={() => setIsEditing(false)}
            onSuccess={(updatedProductType) => {
              setProductType(updatedProductType)
              setIsEditing(false)
            }}
          />
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={isDeleteDialogOpen}
        onOpenChange={(open) => {
          if (!isDeleting) {
            setIsDeleteDialogOpen(open)
            setDeleteError(null)
          }
        }}
      >
      <AlertDialogContent size="sm">
        <AlertDialogHeader>
          <AlertDialogMedia className="bg-destructive/10 text-destructive dark:bg-destructive/20 dark:text-destructive">
            <Trash2Icon />
          </AlertDialogMedia>
          <AlertDialogTitle>Delete Product Type</AlertDialogTitle>
          <AlertDialogDescription>
            {DELETE_CONFIRMATION_TEXT.question}{" "}
            <b>“{productType.name}”</b>? {DELETE_CONFIRMATION_TEXT.warning}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel
            variant="outline"
            disabled={isDeleting}
            onClick={() => {
              setIsDeleteDialogOpen(false)
              setDeleteError(null)
            }}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            disabled={isDeleting}
            onClick={deleteProductType}
            >
            {isDeleting ? "Deleting..." : "Delete"}
          </AlertDialogAction>
        </AlertDialogFooter>

        {deleteError && (
          <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
            {deleteError}
          </div>
        )}

        </AlertDialogContent>
      </AlertDialog>

      <Card className="shadow-sm">
        <CardHeader>
          <CardTitle>General</CardTitle>
          <CardAction>
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <Button
                    variant="secondary"
                    size="icon"
                    aria-label="Product type actions"
                  />
                }
              >
                <EllipsisVertical />
              </DropdownMenuTrigger>

              <DropdownMenuContent>
                <DropdownMenuItem onClick={() => setIsEditing(true)}>
                  <Pencil />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                  variant="destructive"
                  onClick={() => setIsDeleteDialogOpen(true)}
                >
                  <Trash2Icon />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </CardAction>
        </CardHeader>

        <CardContent className="space-y-6">
          <div className="grid gap-6 lg:grid-cols-2">
            <div>
              <p className="text-sm text-muted-foreground">Name</p>
              <p className="mt-1 font-medium">{productType.name}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Slug</p>
              <p className="mt-1">{productType.slug}</p>
            </div>
          </div>

          <div>
            <p className="text-sm text-muted-foreground">Description</p>
            <p className="mt-1 whitespace-pre-wrap">
              {productType.description || "—"}
            </p>
          </div>
        </CardContent>
      </Card>

      <Card className="shadow-sm">
        <CardHeader>
          <CardTitle>Information</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">ID</p>
          <p className="mt-1 break-all font-mono text-sm">
            {productType.id}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
