import * as React from "react"
import { cn } from "@/lib/utils"

function SectionPanel({
  className,
  size = "default",
  ...props
}: React.ComponentProps<"div"> & { size?: "default" | "sm" }) {
  return (
    <div
      data-slot="card"
      data-size={size}
      className={cn(
        "group/card flex flex-col gap-(--card-spacing) overflow-hidden py-(--card-spacing) [--card-spacing:--spacing(4)] has-data-[slot=card-footer]:pb-0 has-[>img:first-child]:pt-0 data-[size=sm]:[--card-spacing:--spacing(3)] data-[size=sm]:has-data-[slot=card-footer]:pb-0 border-b-[1px]",
        className
      )}
      {...props}
    />
  )
}

function SectionPanelHeader({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="card-header"
      className={cn(
        "group/card-header @container/card-header grid auto-rows-min items-start gap-1 rounded-t-xl has-data-[slot=card-action]:grid-cols-[1fr_auto] has-data-[slot=card-description]:grid-rows-[auto_auto] [.border-b]:pb-(--card-spacing)",
        className
      )}
      {...props}
    />
  )
}

function SectionPanelTitle({ className, children, tailButton, ...props }: React.ComponentProps<"div"> & { tailButton?: React.ReactNode }) {
  return (
    <div
      data-slot="card-title"
      className={cn(
        "font-heading text-xl leading-snug font-medium",
        tailButton && "flex items-center justify-between gap-3",
        className
      )}
      {...props}
    >
      {children}
      {tailButton && <div className="shrink-0">{tailButton}</div>}
    </div>
  )
}

function SectionPanelDescription({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="card-description"
      className={cn("text-sm text-muted-foreground", className)}
      {...props}
    />
  )
}

function SectionPanelContent({
  className,
  leftColumn,
  rightColumn,
  children,
  ...props
}: React.ComponentProps<"div"> & {
  leftColumn?: React.ReactNode
  rightColumn?: React.ReactNode
}) {
  return (
    <div data-slot="card-content" className={cn("flex w-full flex-col gap-4", className)} {...props}>
      {(leftColumn != null || rightColumn != null) && (
        <div className="mt-6 grid gap-6 lg:grid-cols-2">
          <div data-slot="card-content-left-column" className="space-y-6">{leftColumn}</div>
          <div data-slot="card-content-right-column" className="space-y-6">{rightColumn}</div>
        </div>
      )}
      {children}
    </div>
  )
}

function SectionPanelFooter({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="card-footer"
      className={cn(
        "flex items-center rounded-b-xl border-t bg-muted/50 p-(--card-spacing)",
        className
      )}
      {...props}
    />
  )
}

export {
  SectionPanel,
  SectionPanelHeader,
  SectionPanelFooter,
  SectionPanelTitle,
  SectionPanelDescription,
  SectionPanelContent,
}
