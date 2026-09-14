"use client"

import { useRef, useState, type ReactElement } from "react"
import { Plus } from "lucide-react"

import { CollectionFormContainer } from "@/components/admin/collection/CollectionFormContainer"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import type { Collection } from "@/types/collection"

interface CollectionDialogProps {
  collection?: Collection
  trigger?: ReactElement
  onSuccess?: (collection: Collection) => void
  open?: boolean
  onOpenChange?: (open: boolean) => void
  showTrigger?: boolean
}

export function CollectionDialog({
  collection,
  trigger,
  onSuccess,
  open: controlledOpen,
  onOpenChange,
  showTrigger = true,
}: CollectionDialogProps) {
  const [internalOpen, setInternalOpen] = useState(false)
  const formContainerRef = useRef<HTMLDivElement>(null)
  const isEditing = collection !== undefined
  const open = controlledOpen ?? internalOpen

  function setOpen(nextOpen: boolean) {
    if (controlledOpen === undefined) {
      setInternalOpen(nextOpen)
    }

    onOpenChange?.(nextOpen)
  }

  function submitForm() {
    formContainerRef.current?.querySelector("form")?.requestSubmit()
  }

  return (
    <Dialog open={open} onOpenChange={setOpen} disablePointerDismissal>
      {showTrigger && (
        <DialogTrigger render={trigger ?? <Button />}>
          {!trigger && (
            <>
              <Plus />
              Add Collection
            </>
          )}
        </DialogTrigger>
      )}

      <DialogContent
        className="max-h-[90vh] overflow-y-auto sm:max-w-3xl"
        showCloseButton={false}
        primaryButtonLabel={isEditing ? "Save Changes" : "Save Collection"}
        onCancel={() => setOpen(false)}
        onSave={submitForm}
      >
        <DialogHeader>
          <DialogTitle>
            {isEditing ? "Edit Collection" : "New Collection"}
          </DialogTitle>
        </DialogHeader>

        <div ref={formContainerRef}>
          <CollectionFormContainer
            collectionId={collection?.id}
            initialValues={collection}
            onSuccess={(savedCollection) => {
              onSuccess?.(savedCollection)
              setOpen(false)
            }}
          />
        </div>
      </DialogContent>
    </Dialog>
  )
}
