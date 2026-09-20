import * as React from "react"
import { Input as InputPrimitive } from "@base-ui/react/input"

function Input({
  className,
  type = "text",
  ...props
}: React.ComponentProps<"input">) {
  return (
    <InputPrimitive
      type={type}
      data-slot="input"
      className={`h-10 w-full rounded-md border border-transparent hover:border-border-main bg-background-main px-3 py-2 text-sm outline-none placeholder:text-text-gray-light focus:border-primary focus:ring-1 focus:ring-primary/50 disabled:cursor-not-allowed disabled:opacity-50 ${className ?? ""}`}
      {...props}
    />
  )
}

export { Input }