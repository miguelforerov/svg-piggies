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
  const formId = "new-product-form"

  function submitProductForm() {
    const form = document.getElementById(formId)

    if (form instanceof HTMLFormElement) {
      form.requestSubmit()
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen} disablePointerDismissal>
      <DialogTrigger className={buttonVariants()}>
        <Plus />
        Add Product
      </DialogTrigger>

      <DialogContent
        className="max-h-[90vh] overflow-y-auto sm:max-w-3xl"
        primaryButtonLabel="Save Product"
        onCancel={() => setOpen(false)}
        onSave={submitProductForm}
      >
        <DialogHeader>
          <DialogTitle>New Product</DialogTitle>
        </DialogHeader>

        <ProductFormContainer formId={formId} onCancel={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  )
}
