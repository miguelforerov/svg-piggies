"use client"

import { useRef, useState, type ReactElement } from "react"
import { Plus } from "lucide-react"

import { ProductTypeFormContainer } from "@/components/admin/product-type/ProductTypeFormContainer"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import type { ProductType } from "@/types/product-type"

interface ProductTypeDialogProps {
  productType?: ProductType
  trigger?: ReactElement
  onSuccess?: (productType: ProductType) => void
}

export function ProductTypeDialog({
  productType,
  trigger,
  onSuccess,
}: ProductTypeDialogProps) {
  const [open, setOpen] = useState(false)
  const formContainerRef = useRef<HTMLDivElement>(null)
  const isEditing = productType !== undefined

  function submitForm() {
    formContainerRef.current?.querySelector("form")?.requestSubmit()
  }

  return (
    <Dialog open={open} onOpenChange={setOpen} disablePointerDismissal>
      <DialogTrigger render={trigger ?? <Button />}>
        {!trigger && (
          <>
            <Plus />
            Add Product Type
          </>
        )}
      </DialogTrigger>

      <DialogContent
        className="max-h-[90vh] overflow-y-auto sm:max-w-3xl"
        showCloseButton={false}
        primaryButtonLabel={isEditing ? "Save Changes" : "Save Product Type"}
        onCancel={() => setOpen(false)}
        onSave={submitForm}
      >
        <DialogHeader>
          <DialogTitle>
            {isEditing ? "Edit Product Type" : "New Product Type"}
          </DialogTitle>
        </DialogHeader>

        <div ref={formContainerRef}>
          <ProductTypeFormContainer
            productTypeId={productType?.id}
            initialValues={productType}
            onCancel={() => setOpen(false)}
            onSuccess={(savedProductType) => {
              onSuccess?.(savedProductType)
              setOpen(false)
            }}
          />
        </div>
      </DialogContent>
    </Dialog>
  )
}
