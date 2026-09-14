import * as React from "react"
import { Dialog as DialogPrimitive } from "@base-ui/react/dialog"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { XIcon } from "lucide-react"

function Dialog({ ...props }: DialogPrimitive.Root.Props) {
  return <DialogPrimitive.Root data-slot="dialog" {...props} />
}

function DialogTrigger({ ...props }: DialogPrimitive.Trigger.Props) {
  return <DialogPrimitive.Trigger data-slot="dialog-trigger" {...props} />
}

function DialogPortal({ ...props }: DialogPrimitive.Portal.Props) {
  return <DialogPrimitive.Portal data-slot="dialog-portal" {...props} />
}

function DialogClose({ ...props }: DialogPrimitive.Close.Props) {
  return <DialogPrimitive.Close data-slot="dialog-close" {...props} />
}

function DialogOverlay({
  className,
  ...props
}: DialogPrimitive.Backdrop.Props) {
  return (
    <DialogPrimitive.Backdrop
      data-slot="dialog-overlay"
      className={cn(
        "fixed inset-0 isolate z-50 bg-black/10 duration-100 supports-backdrop-filter:backdrop-blur-xs data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0",
        className
      )}
      {...props}
    />
  )
}

function DialogContent({
  className,
  children,
  showCloseButton = true,
  showDialogFooter = true,
  primaryButtonLabel = "Save",
  onCancel,
  onSave,
  ...props
}: DialogPrimitive.Popup.Props & {
  showCloseButton?: boolean
  showDialogFooter?: boolean
  primaryButtonLabel?: string
  onCancel?: () => void
  onSave?: () => void
}) {
  return (
    <DialogPortal>
      <DialogOverlay />
      <DialogPrimitive.Popup
        data-slot="dialog-content"
        className={cn(
          "fixed top-1/2 left-1/2 z-50 grid w-full max-w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 gap-4 rounded-xl bg-popover p-4 text-sm text-popover-foreground ring-1 ring-foreground/10 duration-100 outline-none sm:max-w-sm data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95",
          className
        )}
        {...props}
      >
        {children}
        {showCloseButton && (
          <DialogPrimitive.Close
            data-slot="dialog-close"
            render={
              <Button
                variant="ghost"
                className="absolute top-2 right-2"
                size="icon-sm"
              />
            }
          >
            <XIcon
            />
            <span className="sr-only">Close</span>
          </DialogPrimitive.Close>
        )}
        {showDialogFooter && (
          <DialogFooter
            primaryButtonLabel={primaryButtonLabel}
            onCancel={onCancel}
            onSave={onSave}
          />
        )}
      </DialogPrimitive.Popup>
    </DialogPortal>
  )
}

function DialogHeader({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div className="flex items-center">
      <div
        data-slot="dialog-header"
        className={cn("flex flex-col gap-2 flex-auto", className)}
        {...props}
      />
      <div className="h-4 w-4">
        <svg width="100%" height="100%" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path
            fill="black"
            d="M4.575 4.57498C4.85629 4.29378 5.23775 4.1358 5.6355 4.1358C6.03324 4.1358 6.41471 4.29378 6.696 4.57498L12 9.87898L17.304 4.57498C17.5869 4.30175 17.9658 4.15055 18.3591 4.15397C18.7524 4.15739 19.1286 4.31514 19.4067 4.59326C19.6848 4.87137 19.8426 5.24759 19.846 5.64088C19.8494 6.03418 19.6982 6.41308 19.425 6.69598L14.121 12L19.425 17.304C19.6982 17.5869 19.8494 17.9658 19.846 18.3591C19.8426 18.7524 19.6848 19.1286 19.4067 19.4067C19.1286 19.6848 18.7524 19.8426 18.3591 19.846C17.9658 19.8494 17.5869 19.6982 17.304 19.425L12 14.121L6.696 19.425C6.4131 19.6982 6.03419 19.8494 5.6409 19.846C5.2476 19.8426 4.87138 19.6848 4.59327 19.4067C4.31516 19.1286 4.1574 18.7524 4.15399 18.3591C4.15057 17.9658 4.30176 17.5869 4.575 17.304L9.879 12L4.575 6.69598C4.29379 6.41469 4.13582 6.03323 4.13582 5.63548C4.13582 5.23774 4.29379 4.85627 4.575 4.57498Z"
          />
        </svg>
      </div>
    </div>
  )
}

function DialogFooter({
  className,
  children,
  primaryButtonLabel = "Save",
  onCancel,
  onSave,
  ...props
}: React.ComponentProps<"div"> & {
  primaryButtonLabel?: string
  onCancel?: () => void
  onSave?: () => void
}) {
  return (
    <div
      data-slot="dialog-footer"
      className={cn(
        "-mx-4 -mb-4 flex flex-col-reverse gap-4 rounded-b-xl p-4 sm:flex-row sm:justify-end",
        className
      )}
      {...props}
    >
      {children}
      <DialogPrimitive.Close
        render={<Button variant="outline" onClick={onCancel} />}
      >
        Cancel
      </DialogPrimitive.Close>

      <div className="flex items-center justify-end gap-3">
        <Button type="button" onClick={onSave} disabled={!onSave}>
          {primaryButtonLabel}
        </Button>
      </div>
    </div>
  )
}

function DialogTitle({ className, ...props }: DialogPrimitive.Title.Props) {
  return (
    <DialogPrimitive.Title
      data-slot="dialog-title"
      className={cn(
        "font-heading text-base leading-none font-medium",
        className
      )}
      {...props}
    />
  )
}

function DialogDescription({
  className,
  ...props
}: DialogPrimitive.Description.Props) {
  return (
    <DialogPrimitive.Description
      data-slot="dialog-description"
      className={cn(
        "text-sm text-muted-foreground *:[a]:underline *:[a]:underline-offset-3 *:[a]:hover:text-foreground",
        className
      )}
      {...props}
    />
  )
}

export {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
  DialogTrigger,
}
