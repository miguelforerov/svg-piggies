import { useState, type CSSProperties } from "react"
import { ChevronDown, Menu, X } from "lucide-react"

import menu from "@/config/sidebar-links.json"
import { Button } from "@/components/ui/button"
import {
  Sidebar as ShadcnSidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarProvider,
  SidebarSeparator,
  useSidebar,
} from "@/components/ui/sidebar"

type MenuLink = { name: string; url: string }
type MenuSection = MenuLink & { hasChildren?: boolean; children?: MenuLink[] }
type MenuEntry = MenuSection | { type: "separator" }

function isActive(pathname: string, url: string) {
  const currentPath = pathname.replace(/\/+$/, "") || "/"
  const linkPath = url.replace(/\/+$/, "") || "/"
  return currentPath === linkPath
}

function Submenu({ item, pathname }: { item: MenuSection; pathname: string }) {
  const children = item.children ?? []
  const active = children.some((child) => isActive(pathname, child.url))
  const [open, setOpen] = useState(active)

  return (
    <SidebarMenuItem>
      <SidebarMenuButton
        type="button"
        isActive={active}
        aria-expanded={open}
        onClick={() => setOpen((value) => !value)}
      >
        <span>{item.name}</span>
        <ChevronDown className={`ml-auto transition-transform ${open ? "rotate-180" : ""}`} />
      </SidebarMenuButton>
      {open && (
        <SidebarMenuSub>
          {children.map((child) => (
            <SidebarMenuSubItem key={child.url}>
              <SidebarMenuSubButton
                render={<a href={child.url} />}
                isActive={isActive(pathname, child.url)}
              >
                {child.name}
              </SidebarMenuSubButton>
            </SidebarMenuSubItem>
          ))}
        </SidebarMenuSub>
      )}
    </SidebarMenuItem>
  )
}

function SidebarContents({ pathname }: { pathname: string }) {
  const { isMobile, open, openMobile, setOpen, setOpenMobile, toggleSidebar } = useSidebar()

  const closeSidebar = () => {
    if (isMobile) setOpenMobile(false)
    else setOpen(false)
  }

  return (
    <>
      <Button
        type="button"
        variant="link"
        size="icon"
        className="text-primary"
        aria-label="Open menu"
        aria-expanded={isMobile ? openMobile : open}
        onClick={toggleSidebar}
      >
        <Menu className="size-6" />
      </Button>

      {open && !isMobile && (
        <button
          type="button"
          className="fixed inset-0 z-40 bg-black/50"
          aria-label="Close menu"
          onClick={closeSidebar}
        />
      )}

      <ShadcnSidebar side="left" collapsible="offcanvas" className="z-50 shadow-lg">
        <SidebarHeader className="flex-row items-center justify-between p-6">
          <a href="/" aria-label="SVG Piggies home">
            <img src="/logo.png" alt="SVG Piggies" className="size-24 object-contain" />
          </a>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label="Close menu"
            onClick={closeSidebar}
          >
            <X className="size-6 text-muted-foreground" />
          </Button>
        </SidebarHeader>

        <SidebarContent>
          <SidebarGroup className="px-4">
            <SidebarGroupContent>
              <SidebarMenu>
                {(menu.main as MenuEntry[]).map((item, index) => {
                  if ("type" in item) {
                    return (
                      <SidebarMenuItem key={`separator-${index}`} role="separator" className="py-2">
                        <SidebarSeparator className="mx-0" />
                      </SidebarMenuItem>
                    )
                  }

                  if (item.hasChildren) {
                    return <Submenu key={item.name} item={item} pathname={pathname} />
                  }

                  return (
                    <SidebarMenuItem key={item.url}>
                      <SidebarMenuButton
                        render={<a href={item.url} />}
                        isActive={isActive(pathname, item.url)}
                      >
                        {item.name}
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
      </ShadcnSidebar>
    </>
  )
}

export default function StoreSidebar({ pathname }: { pathname: string }) {
  return (
    <SidebarProvider
      defaultOpen={false}
      className="min-h-0! w-auto! [&_[data-slot=sidebar-gap]]:hidden"
      style={{ "--sidebar-width": "24rem" } as CSSProperties}
    >
      <SidebarContents pathname={pathname} />
    </SidebarProvider>
  )
}
