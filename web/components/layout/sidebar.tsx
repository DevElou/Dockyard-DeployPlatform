"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { LayoutGrid, Server, Settings } from "lucide-react";
import { cn } from "@/lib/utils";

const nav = [
  { href: "/projects", label: "Projects", icon: LayoutGrid },
  { href: "/settings/runtime-targets", label: "Runtime Targets", icon: Server },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="flex h-full w-56 flex-col border-r bg-background">
      <div className="flex h-14 items-center border-b px-4">
        <Link href="/projects" className="flex items-center gap-2 font-semibold">
          <span className="text-primary">⚓</span>
          <span>Dockyard</span>
        </Link>
      </div>

      <nav className="flex-1 space-y-1 p-2">
        {nav.map(({ href, label, icon: Icon }) => {
          const active =
            href === "/projects"
              ? pathname === "/projects" || pathname.startsWith("/projects/")
              : pathname === href || pathname.startsWith(href + "/");
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors",
                active
                  ? "bg-accent text-accent-foreground font-medium"
                  : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
              )}
            >
              <Icon className="h-4 w-4 shrink-0" />
              {label}
            </Link>
          );
        })}
      </nav>

      <div className="border-t p-3">
        <p className="text-xs text-muted-foreground">Dockyard v0.1</p>
      </div>
    </aside>
  );
}
