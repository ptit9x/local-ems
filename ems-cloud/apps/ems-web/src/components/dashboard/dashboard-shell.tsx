import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Activity } from "lucide-react";
import { api } from "../../api";
import type {
  Alert,
  DashboardData,
  Device,
  PageKey,
  Site,
  Telemetry,
  TenantKey,
} from "../../dashboard-types";
import { Sidebar } from "../../layouts/sidebar";
import { Header } from "../../layouts/header";
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
        <Header
          alerts={data.alerts}
          setPage={setPage}
          openNav={() => setMobileNav(true)}
          signedOut={signedOut}
        />
        <main className="mx-auto w-full max-w-[1680px] px-4 py-6 sm:px-5 md:px-7 md:py-8">
          {content}
        </main>
      </div>
    </div>
  );
}
