import { Badge } from "@/components/ui/badge"

interface ProductStatusBadgeProps {
  status: "draft" | "active" | "archived"
}

export function ProductStatusBadge({
  status,
}: ProductStatusBadgeProps) {
  const styles = {
    draft:
      "bg-blue-50 text-blue-800 dark:bg-blue-950 dark:text-blue-400",
    active:
      "bg-green-50 text-green-800 dark:bg-green-950 dark:text-green-400",
    archived:
      "bg-muted text-muted-foreground",
  }

  const labels = {
    draft: "Draft",
    active: "Published",
    archived: "Archived",
  }

  return (
    <Badge
      variant="secondary"
      className={styles[status]}
    >
      {labels[status]}
    </Badge>
  )
}