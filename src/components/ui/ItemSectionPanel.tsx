import type { ReactNode } from "react"

interface ItemSectionPanelProps {
  Title: ReactNode
  Value: ReactNode
}

export function ItemSectionPanel({ Title, Value }: ItemSectionPanelProps) {
  return (
    <div>
      <p className="text-sm text-muted-foreground">{Title}</p>
      <p className="mt-1 font-medium">{Value}</p>
    </div>
  )
}
