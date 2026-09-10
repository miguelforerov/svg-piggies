"use client"

import { useState } from "react"
import { Plus } from "lucide-react"
import { CollectionFormContainer } from "@/components/admin/CollectionFormContainer"
import { buttonVariants } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

export function CollectionFormDialog() {
  const [open, setOpen] = useState(false)

  return (
    <Dialog open={open} onOpenChange={setOpen} disablePointerDismissal>
      <DialogTrigger className={buttonVariants()}>
        <Plus />
        Add Collection
      </DialogTrigger>

      <DialogContent
        className="max-h-[90vh] overflow-y-auto sm:max-w-3xl"
        showCloseButton={false}
      >
        <DialogHeader>
          <DialogTitle>New Collection</DialogTitle>
        </DialogHeader>

        <CollectionFormContainer onCancel={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  )
}
