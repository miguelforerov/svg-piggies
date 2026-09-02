import type { LucideIcon } from "lucide-react"

interface StatsCardProps {
  title: string
  value: number
  icon: LucideIcon
  description?: string
}

export function StatsCard({
  title,
  value,
  icon: Icon,
  description,
}: StatsCardProps) {
  return (
    <div className="rounded-xl border bg-card p-6 text-card-foreground shadow-sm">
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium text-muted-foreground">
          {title}
        </p>

        <Icon className="size-4 text-muted-foreground" />
      </div>

      <div className="mt-3">
        <p className="text-3xl font-semibold tracking-tight">
          {value}
        </p>

        {description && (
          <p className="mt-1 text-xs text-muted-foreground">
            {description}
          </p>
        )}
      </div>
    </div>
  )
}