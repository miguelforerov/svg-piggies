"use client"

import type { ReactNode } from "react"
import {
  SidebarProvider,
  SidebarInset,
} from "@/components/ui/sidebar"
import { Header } from "@/components/admin/Header"
import { Sidebar } from "@/components/admin/Sidebar"

interface ShellProps {
  title: string
  children: ReactNode
}

export function Shell({ title, children }: ShellProps) {
  return (
    <SidebarProvider>
      <Sidebar />

      <SidebarInset>
        <Header title={title} />

        <main className="flex flex-1 flex-col gap-6 p-6">
          {children}
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}