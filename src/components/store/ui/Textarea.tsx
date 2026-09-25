import * as React from "react"

function Textarea({
  className,
  ...props
}: React.ComponentProps<"textarea">) {
  return (
    <textarea
      data-slot="textarea"
      className={`min-h-16 w-full resize-y rounded-md border border-transparent bg-background-main px-3 py-2 text-sm outline-none placeholder:text-text-gray-light hover:border-border-main focus:border-primary focus:ring-1 focus:ring-primary/50 disabled:cursor-not-allowed disabled:opacity-50 ${className ?? ""}`}
      {...props}
    />
  )
}

export { Textarea }
