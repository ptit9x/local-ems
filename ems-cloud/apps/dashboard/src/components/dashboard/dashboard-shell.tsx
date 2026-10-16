import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import {
  Activity,
  Bell,
  Building2,
  Check,
  ChevronDown,
  CircleUserRound,
  Command,
  Grid2X2,
  HelpCircle,
  LogOut,
  Menu,
  Network,
  Search,
  Settings,
  ShieldCheck,
  TrendingUp,
  X,
} from "lucide-react";
import { api, logout } from "../../api";
import type {
  Alert,
  DashboardData,
  Device,
  PageKey,
  Site,
  Telemetry,
  TenantKey,
} from "../../dashboard-types";
import { cn } from "../../lib/utils";
import { Badge } from "../ui/badge";
import { Brand as Logo } from "../ui/brand";
import { Button } from "../ui/button";
import {
  AlertsPage,
  CommandsPage,
  DevicesPage,
  EnergyPage,
  HelpPage,
  OverviewPage,
  ProfilePage,
  ReportsPage,
  SettingsPage,
  SitesPage,
} from "./pages";

function NotificationMenu({
  alerts,
  onViewAll,
}: {
  alerts: Alert[];
  onViewAll: () => void;
}) {
  const [read, setRead] = useState(false);
  const active = alerts.filter((item) => item.active);
  const items = active.length
    ? active.slice(0, 3)
    : [
        {
          id: "system-ready",
          message: "All energy services are operating normally",
          severity: "INFO",
          createdAt: new Date().toISOString(),
          active: false,
        },
      ];
  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild>
        <Button
          aria-label="Notifications"
          variant="ghost"
          size="icon"
          className="relative"
        >
          <Bell className="size-[18px]" />
          {!read && active.length > 0 && (
            <span className="absolute right-2 top-2 size-1.5 rounded-full bg-amber-500 ring-2 ring-[#f4f5f0]" />
          )}
        </Button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          sideOffset={8}
          className="z-50 w-[calc(100vw-2rem)] max-w-[380px] overflow-hidden rounded-2xl border border-border bg-white shadow-2xl"
        >
          <div className="flex items-center justify-between border-b border-border px-4 py-3.5">
            <div>
              <div className="text-sm font-semibold">Notifications</div>
              <div className="text-[11px] text-muted-foreground">
                {active.length} active updates
              </div>
            </div>
            <button
              onClick={() => setRead(true)}
              className="text-xs font-medium text-emerald-700 hover:underline"
            >
              Mark all read
            </button>
          </div>
          <div className="max-h-[340px] overflow-y-auto p-2">
            {items.map((item) => (
              <DropdownMenu.Item
                key={item.id}
                onSelect={onViewAll}
                className="flex cursor-pointer gap-3 rounded-xl p-3 outline-none hover:bg-slate-50 focus:bg-slate-50"
              >
                <div
                  className={cn(
                    "mt-0.5 grid size-9 shrink-0 place-items-center rounded-xl",
                    item.active
                      ? "bg-amber-50 text-amber-700"
                      : "bg-emerald-50 text-emerald-700",
                  )}
                >
                  {item.active ? (
                    <Bell className="size-4" />
                  ) : (
                    <Check className="size-4" />
                  )}
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-sm font-medium leading-5 text-slate-900">
                    {item.message}
                  </div>
                  <div className="mt-1 text-[11px] text-muted-foreground">
                    {item.severity} ·{" "}
                    {new Date(item.createdAt).toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </div>
                </div>
                {!read && item.active && (
                  <span className="mt-2 size-2 shrink-0 rounded-full bg-amber-500" />
                )}
              </DropdownMenu.Item>
            ))}
          </div>
          <div className="border-t border-border p-2">
            <DropdownMenu.Item
              onSelect={onViewAll}
              className="cursor-pointer rounded-lg px-3 py-2.5 text-center text-sm font-medium text-emerald-700 outline-none hover:bg-emerald-50"
            >
              View all notifications
            </DropdownMenu.Item>
          </div>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}

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
function Sidebar({
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

export function DashboardShell({ signedOut }: { signedOut: () => void }) {
  const [page, setPage] = useState<PageKey>("overview");
  const [tenant, setTenant] = useState<TenantKey>("tenant-a");
  const [siteId, setSiteId] = useState<string>();
  const [mobileNav, setMobileNav] = useState(false);
  const queryClient = useQueryClient();
  const sites = useQuery({
    queryKey: ["sites"],
    queryFn: () => api<Site[]>("/sites"),
    refetchInterval: 5000,
  });
  const selected = siteId ?? sites.data?.[0]?.id;
  const site = sites.data?.find((item) => item.id === selected);
  const latest = useQuery({
    queryKey: ["latest", selected],
    queryFn: () => api<Telemetry | null>(`/sites/${selected}/telemetry/latest`),
    enabled: !!selected,
    refetchInterval: 5000,
  });
  const history = useQuery({
    queryKey: ["history", selected],
    queryFn: () => api<Telemetry[]>(`/sites/${selected}/telemetry?limit=120`),
    enabled: !!selected,
    refetchInterval: 5000,
  });
  const devices = useQuery({
    queryKey: ["devices", selected],
    queryFn: () => api<Device[]>(`/sites/${selected}/devices`),
    enabled: !!selected,
    refetchInterval: 5000,
  });
  const alerts = useQuery({
    queryKey: ["alerts", selected],
    queryFn: () => api<Alert[]>(`/sites/${selected}/alerts`),
    enabled: !!selected,
    refetchInterval: 10000,
  });
  const data: DashboardData = {
    site,
    sites: sites.data ?? [],
    telemetry: latest.data ?? undefined,
    history: history.data ?? [],
    devices: devices.data ?? [],
    alerts: alerts.data ?? [],
    updatedAt: latest.dataUpdatedAt,
  };
  const alertCount = data.alerts.filter((item) => item.active).length;
  if (sites.isLoading)
    return (
      <div className="grid min-h-screen place-items-center bg-[#f4f5f0]">
        <div className="flex items-center gap-3 text-sm text-slate-500">
          <Activity className="size-5 animate-pulse text-emerald-700" />
          Loading your energy workspace…
        </div>
      </div>
    );
  const content =
    page === "overview" ? (
      <OverviewPage
        data={data}
        refresh={() => queryClient.invalidateQueries()}
        selectSite={setSiteId}
        selectedSite={selected}
      />
    ) : page === "energy" ? (
      <EnergyPage data={data} />
    ) : page === "sites" ? (
      <SitesPage data={data} tenant={tenant} />
    ) : page === "devices" ? (
      <DevicesPage data={data} />
    ) : page === "commands" ? (
      <CommandsPage />
    ) : page === "alerts" ? (
      <AlertsPage data={data} />
    ) : page === "reports" ? (
      <ReportsPage />
    ) : page === "profile" ? (
      <ProfilePage />
    ) : page === "help" ? (
      <HelpPage />
    ) : (
      <SettingsPage tenant={tenant} />
    );
  return (
    <div className="min-h-screen bg-[#f4f5f0] text-foreground">
      <Sidebar
        open={mobileNav}
        close={() => setMobileNav(false)}
        page={page}
        setPage={setPage}
        tenant={tenant}
        setTenant={setTenant}
        alertCount={alertCount}
      />
      <div className="lg:pl-[268px]">
        <header className="sticky top-0 z-30 flex h-[68px] items-center border-b border-slate-200/80 bg-[#f4f5f0]/90 px-4 backdrop-blur-xl md:h-[72px] md:px-7">
          <Button
            aria-label="Open navigation"
            variant="ghost"
            size="icon"
            className="mr-2 lg:hidden"
            onClick={() => setMobileNav(true)}
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
              alerts={data.alerts}
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
        <main className="mx-auto w-full max-w-[1680px] px-4 py-6 sm:px-5 md:px-7 md:py-8">
          {content}
        </main>
      </div>
    </div>
  );
}
