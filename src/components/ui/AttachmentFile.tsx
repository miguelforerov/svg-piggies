import { useRef, useState, type ReactNode } from "react"
import { FileUp } from "lucide-react"
import { Attachment, AttachmentAction, AttachmentActions, AttachmentTrigger } from "@/components/ui/attachment"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

type QueuedAttachment = { id: number | string; file: File; stage: string; error?: string }

type QueueAction<T> = {
  onClick: (item: T) => void
  show?: (item: T) => boolean
  disabled?: (item: T) => boolean
}

type AttachmentFileProps<T extends QueuedAttachment> = {
  accept: string
  buttonLabel: string
  description: string
  disabled?: boolean
  icon: ReactNode
  onFiles: (files: FileList) => void
  onUpload: () => void
  uploadDisabled: boolean
  queue: readonly T[]
  statusLabel: (item: T) => string
  sizeLabel?: (item: T) => string
  preview?: (item: T) => string | null | undefined
  previewFallback?: string
  alternativeText?: {
    value: (item: T) => string
    onChange: (item: T, value: string) => void
    show?: (item: T) => boolean
  } | null
  primary?: {
    value: (item: T) => boolean | null
    onClick: (item: T) => void
    disabled?: (item: T) => boolean
    selectedLabel?: string
  } | null
  retry?: QueueAction<T> | null
  remove?: QueueAction<T> | null
  queueItemsClassName?: string
  className?: string
}

function AttachmentFile<T extends QueuedAttachment>({
  accept,
  buttonLabel,
  description,
  disabled = false,
  icon,
  onFiles,
  onUpload,
  uploadDisabled,
  queue,
  statusLabel,
  sizeLabel,
  preview,
  previewFallback,
  alternativeText,
  primary,
  retry,
  remove,
  queueItemsClassName,
  className,
}: AttachmentFileProps<T>) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)

  return (
    <Attachment
        state="idle"
        className={cn(
          "w-full flex-col gap-8 rounded-lg border-2 p-8",
          dragging ? "border-primary bg-muted" : "border-border",
          className
        )}
        onDragOver={(event) => {
          event.preventDefault()
          if (!disabled) setDragging(true)
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={(event) => {
          event.preventDefault()
          setDragging(false)
          if (!disabled && event.dataTransfer.files.length > 0) onFiles(event.dataTransfer.files)
        }}
      >
        <div className="flex flex-col items-center gap-4">
          {icon}
          <p className="text-sm">{description}</p>
          <input
            ref={inputRef}
            type="file"
            accept={accept}
            multiple
            disabled={disabled}
            hidden
            onChange={(event) => {
              if (event.target.files?.length) onFiles(event.target.files)
              event.target.value = ""
            }}
          />
          <AttachmentActions>
            <AttachmentAction type="button" variant="outline" size="sm" disabled={disabled} onClick={() => inputRef.current?.click()}>
              {buttonLabel}
            </AttachmentAction>
          </AttachmentActions>
          <AttachmentTrigger aria-label={buttonLabel} disabled={disabled} onClick={() => inputRef.current?.click()} />
        </div>
        {queue.length > 0 && (
          <div className="flex flex-col relative z-20 w-full gap-8">
            <div className={queueItemsClassName ?? "grid gap-4 sm:grid-cols-2 lg:grid-cols-3 gap-4 xl:gap-8"}>
              {queue.map((item) => {
                const size = sizeLabel?.(item)
                const previewUrl = preview?.(item) || previewFallback
                const isPrimary = primary?.value(item)
                const showAlternativeText = alternativeText && (alternativeText.show?.(item) ?? true)
                const showRetry = retry && (retry.show?.(item) ?? true)
                const showRemove = remove && (remove.show?.(item) ?? true)
                return (
                  <Card key={item.id} className="w-full">
                    <CardContent>
                      <div className="flex flex-wrap items-start gap-3 sm:flex-nowrap">
                        {previewUrl ? (
                          <img
                            src={previewUrl}
                            alt="Selected preview"
                            className="size-16 rounded object-cover"
                            onError={(event) => {
                              if (previewFallback && event.currentTarget.getAttribute("src") !== previewFallback) {
                                event.currentTarget.src = previewFallback
                              }
                            }}
                          />
                        ) : (
                          <FileUp className="size-16 shrink-0 text-muted-foreground" />
                        )}
                        <div className="min-w-0 flex-1 space-y-1">
                          <p className="truncate text-sm font-medium" title={item.file.name}>{item.file.name}</p>
                          <p className="text-xs text-muted-foreground" role="status">
                            {statusLabel(item)}{size && ` · ${size}`}
                          </p>
                          {item.error && <p className="text-xs text-destructive" role="alert">{item.error}</p>}
                          {alternativeText && showAlternativeText && (
                            <Input
                              value={alternativeText.value(item)}
                              placeholder="Alternative text (optional)"
                              aria-label={`Alternative text for ${item.file.name}`}
                              onChange={(event) => alternativeText.onChange(item, event.target.value)}
                            />
                          )}
                        </div>
                        {retry && showRetry && (
                          <Button type="button" variant="outline" size="sm" disabled={retry.disabled?.(item)} onClick={() => retry.onClick(item)}>
                            Retry
                          </Button>
                        )}
                      </div>
                      {((primary && isPrimary != null) || showRemove) && (
                        <div className="flex flex-wrap justify-center items-center gap-4">
                          {primary && isPrimary != null && (
                            isPrimary ?
                              <Badge variant="secondary">{primary.selectedLabel ?? "Primary"}</Badge>
                            : (
                              <Button type="button" variant="outline" size="sm" disabled={primary.disabled?.(item)}
                                aria-label={`Set ${item.file.name} as primary`} onClick={() => primary.onClick(item)}>
                                Set as primary
                              </Button>
                            )
                          )}
                          {remove && showRemove && (
                            <Button type="button" variant="ghost" size="sm" disabled={remove.disabled?.(item)}
                              onClick={() => remove.onClick(item)} aria-label={`Remove ${item.file.name} from selection`}>
                              Remove
                            </Button>
                          )}
                        </div>
                      )}
                    </CardContent>
                  </Card>
                )
              })}
            </div>
            <div className="flex justify-center">
              <Button type="button" className="px-8" disabled={uploadDisabled} onClick={onUpload}>Upload</Button>
            </div>
          </div>
        )}
    </Attachment>
  )
}

export { AttachmentFile }
