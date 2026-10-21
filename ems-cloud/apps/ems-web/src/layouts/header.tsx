import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import {
  ChevronDown,
  CircleUserRound,
  HelpCircle,
  LogOut,
  Menu,
  Search,
  Settings,
} from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import type { Alert, PageKey } from "../dashboard-types";
import { logout } from "../api";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { NotificationMenu } from "./notification-menu";

export function Header({
  alerts,
  setPage,
  openNav,
  signedOut,
}: {
  alerts: Alert[];
  setPage: (page: PageKey) => void;
  openNav: () => void;
  signedOut: () => void;
}) {
  const queryClient = useQueryClient();
  return (
    <header className="sticky top-0 z-30 flex h-[68px] items-center border-b border-slate-200/80 bg-[#f4f5f0]/90 px-4 backdrop-blur-xl md:h-[72px] md:px-7">
      <Button
        aria-label="Open navigation"
        variant="ghost"
        size="icon"
        className="mr-2 lg:hidden"
        onClick={openNav}
      >
        <Menu className="size-5" />
      </Button>
      <div className="hidden max-w-sm flex-1 items-center md:flex">
        <Search className="mr-2 size-4 text-slate-400" />
        <input
          aria-label="Search workspace"
          placeholder="Search sites, devices, commands…"
          className="w-full bg-transparent text-sm outline-none placeholder:text-slate-400"
        />
      </div>
      <div className="ml-auto flex items-center gap-1 sm:gap-2">
        <Badge className="hidden border-emerald-200 bg-emerald-50 text-emerald-700 sm:flex">
          <span className="mr-1.5 size-1.5 rounded-full bg-emerald-500" />
          All systems operational
        </Badge>
        <NotificationMenu
          alerts={alerts}
          onViewAll={() => setPage("alerts")}
        />
        <DropdownMenu.Root>
          <DropdownMenu.Trigger asChild>
            <Button aria-label="Open user menu" variant="ghost" className="h-10 gap-2 px-2">
              <div className="grid size-8 place-items-center rounded-lg bg-[#0c7664] text-white">
                <CircleUserRound className="size-4" />
              </div>
              <div className="hidden text-left md:block">
                <div className="text-xs font-semibold">Demo Admin</div>
                <div className="text-[10px] text-muted-foreground">
                  Administrator
                </div>
              </div>
              <ChevronDown className="size-3 text-muted-foreground" />
            </Button>
          </DropdownMenu.Trigger>
          <DropdownMenu.Portal>
            <DropdownMenu.Content
              align="end"
              className="z-50 min-w-48 rounded-lg border border-border bg-white p-1 shadow-xl"
            >
              <DropdownMenu.Item
                onSelect={() => setPage("profile")}
                className="flex cursor-pointer items-center gap-2 rounded-md px-3 py-2 text-sm outline-none hover:bg-muted"
              >
                <Settings className="size-4" />
                Profile settings
              </DropdownMenu.Item>
              <DropdownMenu.Item
                onSelect={() => setPage("help")}
                className="flex cursor-pointer items-center gap-2 rounded-md px-3 py-2 text-sm outline-none hover:bg-muted"
              >
                <HelpCircle className="size-4" />
                Help center
              </DropdownMenu.Item>
              <DropdownMenu.Separator className="my-1 h-px bg-border" />
              <DropdownMenu.Item
                onSelect={async () => {
                  await logout();
                  queryClient.clear();
                  signedOut();
                }}
                className="flex cursor-pointer items-center gap-2 rounded-md px-3 py-2 text-sm text-red-600 outline-none hover:bg-red-50"
              >
                <LogOut className="size-4" />
                Sign out
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Portal>
        </DropdownMenu.Root>
      </div>
    </header>
  );
}
