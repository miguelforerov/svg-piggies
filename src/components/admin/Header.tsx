import { SidebarTrigger } from "@/components/ui/sidebar"

interface HeaderProps {
  title: string
}

export function Header({ title }: HeaderProps) {
  return (
    <header className="flex h-14 shrink-0 items-center gap-2 border-b bg-background px-4">
      <SidebarTrigger />

      <div className="h-5 w-px bg-border" />

      <h1 className="text-sm font-medium">
        {title}
      </h1>
    </header>
  )
}