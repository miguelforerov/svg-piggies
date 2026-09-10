"use client"

import { useState } from "react"
import { Plus } from "lucide-react"
import { ProductTypeFormContainer } from "@/components/admin/ProductTypeFormContainer"
import { buttonVariants } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

export function ProductTypeFormDialog() {
  const [open, setOpen] = useState(false)

  return (
    <Dialog open={open} onOpenChange={setOpen} disablePointerDismissal>
      <DialogTrigger className={buttonVariants()}>
        <Plus />
        Add Product Type
      </DialogTrigger>

      <DialogContent
        className="max-h-[90vh] overflow-y-auto sm:max-w-3xl"
        showCloseButton={false}
      >
        <DialogHeader>
          <DialogTitle>New Product Type</DialogTitle>
        </DialogHeader>

        <ProductTypeFormContainer onCancel={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  )
}
