"use client"

import { useState } from "react"
import { Plus } from "lucide-react"
import { buttonVariants } from "@/components/ui/button"
import { ProductFormContainer } from "@/components/admin/ProductFormContainer"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

export function ProductFormDialog() {
  const [open, setOpen] = useState(false)

  return (
    <Dialog open={open} onOpenChange={setOpen} disablePointerDismissal>
      <DialogTrigger
        className={buttonVariants()}
      >
        <Plus />
        Add Product
      </DialogTrigger>

      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-3xl" showCloseButton={false}>
        <DialogHeader>
          <DialogTitle>New Product</DialogTitle>
        </DialogHeader>

        <ProductFormContainer onCancel={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  )
}