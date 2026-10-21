import { useState } from "react";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import {
  Activity,
  Bell,
  Building2,
  Check,
  ChevronDown,
  Command,
  Grid2X2,
  Network,
  Settings,
  ShieldCheck,
  TrendingUp,
  X,
} from "lucide-react";
import type { PageKey, TenantKey } from "../dashboard-types";
import { cn } from "../lib/utils";
import { Brand as Logo } from "../components/ui/brand";
import { Button } from "../components/ui/button";

const tenants = [
  {
    id: "tenant-a" as const,
    name: "Demo Tenant A",
    detail: "Live simulator",
    tone: "bg-emerald-300/10 text-emerald-300",
  },
  {
    id: "tenant-b" as const,
    name: "Demo Tenant B",
    detail: "Preview dataset",
    tone: "bg-violet-300/10 text-violet-300",
  },
];

const navigation = [
  { id: "overview" as const, label: "Overview", icon: Grid2X2 },
  { id: "energy" as const, label: "Live energy", icon: Activity },
  { id: "sites" as const, label: "Sites", icon: Building2 },
  { id: "devices" as const, label: "Devices", icon: Network },
  { id: "commands" as const, label: "Commands", icon: Command },
];

const operations = [
  { id: "alerts" as const, label: "Alerts", icon: Bell },
  { id: "reports" as const, label: "Reports", icon: TrendingUp },
  { id: "settings" as const, label: "Settings", icon: Settings },
];

function NavGroup({
  label,
  items,
  page,
  navigate,
  alertCount = 0,
}: {
  label: string;
  items: { id: PageKey; label: string; icon: typeof Activity }[];
  page: PageKey;
  navigate: (page: PageKey) => void;
  alertCount?: number;
}) {
  return (
    <div>
      <p className="px-3 text-[10px] font-semibold uppercase tracking-[.16em] text-slate-600">
        {label}
      </p>
      <div className="mt-2 space-y-1">
        {items.map(({ id, label: itemLabel, icon: Icon }) => (
          <button
            key={id}
            onClick={() => navigate(id)}
            className={cn(
              "flex h-10 w-full items-center gap-3 rounded-lg px-3 text-sm transition",
              page === id
                ? "bg-emerald-300/10 font-medium text-emerald-200"
                : "text-slate-400 hover:bg-white/5 hover:text-white",
            )}
          >
            <Icon className="size-[18px]" />
            {itemLabel}
            {id === "alerts" && alertCount > 0 ? (
              <span className="ml-auto rounded-full bg-amber-300/10 px-2 py-0.5 text-[10px] font-semibold text-amber-300">
                {alertCount}
              </span>
            ) : (
              page === id && (
                <span className="ml-auto size-1.5 rounded-full bg-emerald-300" />
              )
            )}
          </button>
        ))}
      </div>
    </div>
  );
}

export function Sidebar({
  open,
  close,
  page,
  setPage,
  tenant,
  setTenant,
  alertCount,
}: {
  open: boolean;
  close: () => void;
  page: PageKey;
  setPage: (page: PageKey) => void;
  tenant: TenantKey;
  setTenant: (tenant: TenantKey) => void;
  alertCount: number;
}) {
  const current = tenants.find((item) => item.id === tenant)!;
  const navigate = (target: PageKey) => {
    setPage(target);
    close();
    window.scrollTo({ top: 0, behavior: "smooth" });
  };
  return (
    <>
      <button
        aria-label="Close menu"
        onClick={close}
        className={cn(
          "fixed inset-0 z-40 bg-slate-950/40 backdrop-blur-sm lg:hidden",
          open ? "block" : "hidden",
        )}
      />
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-50 flex w-[268px] flex-col overflow-y-auto border-r border-white/10 bg-[#091512] px-4 py-5 text-slate-200 transition-transform lg:translate-x-0",
          open ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <div className="flex items-center justify-between px-2">
          <Logo />
          <Button
            aria-label="Close navigation"
            variant="ghost"
            size="icon"
            onClick={close}
            className="lg:hidden"
          >
            <X className="size-5" />
          </Button>
        </div>
        <DropdownMenu.Root>
          <DropdownMenu.Trigger asChild>
            <button className="mt-8 flex w-full items-center gap-3 rounded-xl border border-white/10 bg-white/[.045] p-3 text-left outline-none transition hover:bg-white/[.07] focus:ring-2 focus:ring-emerald-300/30">
              <div
                className={cn(
                  "grid size-9 place-items-center rounded-lg",
                  current.tone,
                )}
              >
                <Building2 className="size-4" />
              </div>
              <div className="min-w-0 flex-1">
                <div className="truncate text-sm font-medium text-white">
                  {current.name}
                </div>
                <div className="text-[11px] text-slate-500">
                  {current.detail}
                </div>
              </div>
              <ChevronDown className="size-4 text-slate-500" />
            </button>
          </DropdownMenu.Trigger>
          <DropdownMenu.Portal>
            <DropdownMenu.Content
              sideOffset={8}
              align="start"
              className="z-[60] w-[236px] rounded-xl border border-slate-200 bg-white p-1.5 shadow-2xl"
            >
              {tenants.map((item) => (
                <DropdownMenu.Item
                  key={item.id}
                  onSelect={() => setTenant(item.id)}
                  className="flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2.5 outline-none hover:bg-slate-50 focus:bg-slate-50"
                >
                  <div
                    className={cn(
                      "grid size-8 place-items-center rounded-lg",
                      item.id === "tenant-a"
                        ? "bg-emerald-50 text-emerald-700"
                        : "bg-violet-50 text-violet-700",
                    )}
                  >
                    <Building2 className="size-4" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="text-sm font-medium text-slate-900">
                      {item.name}
                    </div>
                    <div className="text-[10px] text-slate-500">
                      {item.detail}
                    </div>
                  </div>
                  {tenant === item.id && (
                    <Check className="size-4 text-emerald-700" />
                  )}
                </DropdownMenu.Item>
              ))}
            </DropdownMenu.Content>
          </DropdownMenu.Portal>
        </DropdownMenu.Root>
        <nav className="mt-7 space-y-6">
          <NavGroup
            label="Workspace"
            items={navigation}
            page={page}
            navigate={navigate}
          />
          <NavGroup
            label="Operations"
            items={operations}
            page={page}
            navigate={navigate}
            alertCount={alertCount}
          />
        </nav>
        <div className="mt-auto pt-8">
          <div className="rounded-xl border border-white/10 bg-gradient-to-br from-emerald-300/10 to-cyan-300/5 p-4">
            <div className="flex items-center gap-2 text-xs font-medium text-emerald-200">
              <ShieldCheck className="size-4" />
              System protected
            </div>
            <p className="mt-2 text-xs leading-5 text-slate-500">
              Tenant-isolated access and encrypted command transport.
            </p>
          </div>
        </div>
      </aside>
    </>
  );
}
