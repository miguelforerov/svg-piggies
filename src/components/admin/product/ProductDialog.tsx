"use client"

import { useId, useState, type ReactElement } from "react"
import { Plus } from "lucide-react"

import {
  ProductFormContainer,
  type Product,
} from "@/components/admin/product/ProductFormContainer"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

interface ProductDialogProps {
  product?: Product
  trigger?: ReactElement
  onSuccess?: (product: Product) => void
  open?: boolean
  onOpenChange?: (open: boolean) => void
  showTrigger?: boolean
}

export function ProductDialog({
  product,
  trigger,
  onSuccess,
  open: controlledOpen,
  onOpenChange,
  showTrigger = true,
}: ProductDialogProps) {
  const [internalOpen, setInternalOpen] = useState(false)
  const formId = `product-form-${useId()}`
  const isEditing = product !== undefined
  const open = controlledOpen ?? internalOpen

  function setOpen(nextOpen: boolean) {
    if (controlledOpen === undefined) {
      setInternalOpen(nextOpen)
    }

    onOpenChange?.(nextOpen)
  }

  function submitProductForm() {
    const form = document.getElementById(formId)

    if (form instanceof HTMLFormElement) {
      form.requestSubmit()
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen} disablePointerDismissal>
      {showTrigger && (
        <DialogTrigger render={trigger ?? <Button />}>
          {!trigger && (
            <>
              <Plus />
              Add Product
            </>
          )}
        </DialogTrigger>
      )}

      <DialogContent
        className="max-h-[90vh] overflow-y-auto sm:max-w-3xl"
        showCloseButton={false}
        primaryButtonLabel={isEditing ? "Save Changes" : "Save Product"}
        onCancel={() => setOpen(false)}
        onSave={submitProductForm}
      >
        <DialogHeader>
          <DialogTitle>
            {isEditing ? "Edit Product" : "New Product"}
          </DialogTitle>
        </DialogHeader>

        <ProductFormContainer
          formId={formId}
          productId={product?.id}
          initialValues={product}
          onSuccess={(savedProduct) => {
            onSuccess?.(savedProduct)
            setOpen(false)
          }}
        />
      </DialogContent>
    </Dialog>
  )
}
