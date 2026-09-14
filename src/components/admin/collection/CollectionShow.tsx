"use client"

import { useEffect, useState } from "react"
import { EllipsisVertical, Pencil, Trash2 } from "lucide-react"
import type { Collection } from "@/types/collection"
import { CollectionDialog } from "@/components/admin/collection/CollectionDialog"
import { Button } from "@/components/ui/button"
import { DELETE_CONFIRMATION_TEXT } from "@/lib/constants"
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogMedia,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface CollectionShowProps {
  collectionId: string
}

export function CollectionShow({
  collectionId,
}: CollectionShowProps) {
  const [collection, setCollection] = useState<Collection | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  useEffect(() => {
    async function loadCollection() {
      try {
        const response = await fetch(
          `/api/admin/collections/${collectionId}`
        )

        if (!response.ok) {
          const message = await response.text()
          throw new Error(message || "Failed to load collection")
        }

        const data: Collection = await response.json()
        setCollection(data)
      } catch (error) {
        console.error("Error loading collection:", error)
        setError(
          error instanceof Error
            ? error.message
            : "Failed to load collection"
        )
      } finally {
        setIsLoading(false)
      }
    }

    loadCollection()
  }, [collectionId])

  async function deleteCollection() {
    setIsDeleting(true)
    setDeleteError(null)

    try {
      const response = await fetch(
        `/api/admin/collections/${collectionId}`,
        { method: "DELETE" }
      )

      if (!response.ok) {
        const message = await response.text()
        throw new Error(message || "Failed to delete collection")
      }

      window.location.href = "/admin/collections"
    } catch (error) {
      console.error("Error deleting collection:", error)
      setDeleteError(
        error instanceof Error
          ? error.message
          : "Failed to delete collection"
      )
      setIsDeleting(false)
    }
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-10 text-center text-muted-foreground">
          Loading collection...
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

  if (!collection) {
    return (
      <Card>
        <CardContent className="p-10 text-center text-muted-foreground">
          Collection not found.
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold tracking-tight">
          {collection.name}
        </h2>
      </div>

      <CollectionDialog
        collection={collection}
        open={isEditing}
        onOpenChange={setIsEditing}
        showTrigger={false}
        onSuccess={setCollection}
      />

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
              <Trash2 />
            </AlertDialogMedia>
            <AlertDialogTitle>Delete Collection</AlertDialogTitle>
            <AlertDialogDescription>
              {DELETE_CONFIRMATION_TEXT.question}{" "}
              <b>“{collection.name}”</b>? {DELETE_CONFIRMATION_TEXT.warning}
            </AlertDialogDescription>
          </AlertDialogHeader>

          {deleteError && (
            <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
              {deleteError}
            </div>
          )}

          <AlertDialogFooter>
            <AlertDialogCancel
              disabled={isDeleting}
              onClick={() => {
                setIsDeleteDialogOpen(false)
                setDeleteError(null)
              }}
            >
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={isDeleting}
              onClick={deleteCollection}
            >
              <Trash2 />
              {isDeleting ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
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
                    aria-label="Collection actions"
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
                  <Trash2 />
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
              <p className="mt-1 font-medium">{collection.name}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Slug</p>
              <p className="mt-1">{collection.slug}</p>
            </div>
          </div>

          <div>
            <p className="text-sm text-muted-foreground">Description</p>
            <p className="mt-1 whitespace-pre-wrap">
              {collection.description || "—"}
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
            {collection.id}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
