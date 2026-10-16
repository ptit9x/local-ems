import { useMemo } from "react";
import ReactEChartsCore from "echarts-for-react/lib/core";
import * as echarts from "echarts/core";
import { BarChart, LineChart } from "echarts/charts";
import {
  GridComponent,
  LegendComponent,
  TooltipComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import {
  Activity,
  AlertTriangle,
  ArrowDownRight,
  ArrowUpRight,
  BatteryCharging,
  BookOpen,
  Building2,
  CalendarDays,
  CheckCircle2,
  ChevronRight,
  Cloud,
  Command,
  Download,
  Filter,
  Gauge,
  KeyRound,
  Leaf,
  Mail,
  MapPin,
  MessageCircle,
  MoreHorizontal,
  Network,
  Play,
  Plus,
  Radio,
  Search,
  Settings2,
  ShieldCheck,
  Sun,
  Users,
  Zap,
} from "lucide-react";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import type {
  DashboardData,
  Device,
  PageKey,
  TenantKey,
  Telemetry,
} from "../../dashboard-types";
import { cn } from "../../lib/utils";

echarts.use([
  LineChart,
  BarChart,
  GridComponent,
  LegendComponent,
  TooltipComponent,
  CanvasRenderer,
]);

export function formatPower(watts = 0) {
  return Math.abs(watts) >= 1000
    ? `${(watts / 1000).toFixed(1)} kW`
    : `${Math.round(watts)} W`;
}

const pageMeta: Record<
  PageKey,
  { eyebrow: string; title: string; description: string }
> = {
  overview: {
    eyebrow: "Fleet operations / Overview",
    title: "Good morning, Richard.",
    description: "Here’s what’s happening across your energy fleet today.",
  },
  energy: {
    eyebrow: "Operations / Live energy",
    title: "Live energy",
    description: "Follow every watt across generation, storage, load and grid.",
  },
  sites: {
    eyebrow: "Portfolio / Sites",
    title: "Sites",
    description:
      "Manage the health and performance of your distributed portfolio.",
  },
  devices: {
    eyebrow: "Assets / Devices",
    title: "Connected devices",
    description:
      "Monitor connectivity, firmware and device health in one view.",
  },
  commands: {
    eyebrow: "Operations / Commands",
    title: "Command center",
    description: "Issue, track and audit remote control actions safely.",
  },
  alerts: {
    eyebrow: "Operations / Alerts",
    title: "Alerts",
    description: "Prioritize active issues before they impact production.",
  },
  reports: {
    eyebrow: "Intelligence / Reports",
    title: "Reports",
    description: "Turn operational data into clear performance insights.",
  },
  settings: {
    eyebrow: "Workspace / Settings",
    title: "Settings",
    description: "Configure your tenant, notifications and platform defaults.",
  },
  profile: {
    eyebrow: "Account / Profile",
    title: "Profile settings",
    description:
      "Manage your identity, account security and personal preferences.",
  },
  help: {
    eyebrow: "Support / Help center",
    title: "How can we help?",
    description:
      "Find answers, operating guides and direct support for CDI Energy.",
  },
};

export function PageHeader({
  page,
  actions,
}: {
  page: PageKey;
  actions?: React.ReactNode;
}) {
  const meta = pageMeta[page];
  return (
    <div className="flex flex-col justify-between gap-5 md:flex-row md:items-end">
      <div className="min-w-0">
        <div className="flex items-center gap-2 text-xs font-medium text-emerald-700">
          <Cloud className="size-3.5" />
          {meta.eyebrow}
        </div>
        <h1 className="mt-2 font-display text-3xl font-semibold tracking-[-.035em] md:text-4xl">
          {meta.title}
        </h1>
        <p className="mt-2 max-w-2xl text-sm text-muted-foreground">
          {meta.description}
        </p>
      </div>
      {actions && (
        <div className="flex shrink-0 flex-wrap items-center gap-2">
          {actions}
        </div>
      )}
    </div>
  );
}

function MetricCard({
  label,
  value,
  unit,
  delta,
  icon: Icon,
  tone,
}: {
  label: string;
  value: string;
  unit: string;
  delta: string;
  icon: typeof Sun;
  tone: string;
}) {
  return (
    <Card className="metric-card overflow-hidden">
      <CardContent className="p-5">
        <div className="flex items-start justify-between">
          <div
            className={cn("grid size-10 place-items-center rounded-xl", tone)}
          >
            <Icon className="size-[19px]" />
          </div>
          <Badge className="border-emerald-100 bg-emerald-50 text-emerald-700">
            {delta}
          </Badge>
        </div>
        <div className="mt-5">
          <p className="text-sm text-muted-foreground">{label}</p>
          <div className="mt-1 flex items-end gap-1.5">
            <span className="font-display text-[30px] font-semibold tracking-[-.04em]">
              {value}
            </span>
            <span className="mb-1 text-xs font-medium text-muted-foreground">
              {unit}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function EnergyFlow({ telemetry }: { telemetry?: Telemetry }) {
  const nodes = [
    {
      label: "Solar",
      value: telemetry?.avgSolarW ?? 0,
      icon: Sun,
      className: "left-1/2 top-0 -translate-x-1/2 text-amber-600",
    },
    {
      label: "Grid",
      value: telemetry?.avgGridW ?? 0,
      icon: Zap,
      className: "left-0 top-1/2 -translate-y-1/2 text-cyan-700",
    },
    {
      label: "Battery",
      value: telemetry?.avgEssW ?? 0,
      icon: BatteryCharging,
      className: "right-0 top-1/2 -translate-y-1/2 text-emerald-700",
    },
    {
      label: "Site load",
      value: telemetry?.avgLoadW ?? 0,
      icon: Building2,
      className: "bottom-0 left-1/2 -translate-x-1/2 text-violet-700",
    },
  ];
  return (
    <div className="relative mx-auto h-[270px] max-w-[420px] sm:h-[285px]">
      <div className="energy-lines absolute inset-[54px]" />
      <div className="absolute left-1/2 top-1/2 grid size-20 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full border border-emerald-100 bg-white shadow-[0_15px_45px_-20px_rgba(5,120,96,.5)] sm:size-24">
        <div className="text-center">
          <Leaf className="mx-auto size-5 text-emerald-600" />
          <div className="mt-1 text-[9px] font-semibold uppercase tracking-wider text-slate-400 sm:text-[10px]">
            Live EMS
          </div>
        </div>
      </div>
      {nodes.map(({ label, value, icon: Icon, className }) => (
        <div
          key={label}
          className={cn(
            "absolute flex w-24 flex-col items-center rounded-xl border border-slate-100 bg-white px-2 py-2 shadow-sm sm:w-28 sm:px-3 sm:py-2.5",
            className,
          )}
        >
          <Icon className="size-4" />
          <span className="mt-1 text-[9px] font-medium text-slate-400 sm:text-[10px]">
            {label}
          </span>
          <span className="font-display text-xs font-semibold text-slate-900 sm:text-sm">
            {formatPower(value)}
          </span>
        </div>
      ))}
    </div>
  );
}

function useEnergyChart(history: Telemetry[]) {
  return useMemo(
    () => ({
      animationDuration: 500,
      color: ["#e5a11a", "#7557d4", "#0f766e", "#1593a8"],
      tooltip: {
        trigger: "axis",
        backgroundColor: "#091512",
        borderWidth: 0,
        textStyle: { color: "#fff" },
      },
      legend: {
        right: 0,
        top: 0,
        itemWidth: 8,
        itemHeight: 8,
        textStyle: { color: "#667085", fontSize: 11 },
      },
      grid: { left: 4, right: 12, bottom: 4, top: 45, containLabel: true },
      xAxis: {
        type: "category",
        boundaryGap: false,
        data: history.map((item) =>
          new Date(item.timestamp).toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
          }),
        ),
        axisLine: { lineStyle: { color: "#e7e9e4" } },
        axisTick: { show: false },
        axisLabel: { color: "#98a2a0", fontSize: 10, hideOverlap: true },
      },
      yAxis: {
        type: "value",
        splitLine: { lineStyle: { color: "#eef0eb", type: "dashed" } },
        axisLabel: {
          color: "#98a2a0",
          fontSize: 10,
          formatter: (value: number) => `${value / 1000}k`,
        },
      },
      series: (
        [
          ["Solar", "avgSolarW"],
          ["Load", "avgLoadW"],
          ["ESS", "avgEssW"],
          ["Grid", "avgGridW"],
        ] as const
      ).map(([name, key]) => ({
        name,
        type: "line",
        smooth: true,
        symbol: "none",
        lineStyle: { width: 2 },
        areaStyle: { opacity: name === "Solar" ? 0.08 : 0 },
        data: history.map((item) => item[key]),
      })),
    }),
    [history],
  );
}

function DeviceTable({
  devices,
  siteCode,
  limit,
}: {
  devices: Device[];
  siteCode?: string;
  limit?: number;
}) {
  const rows = limit ? devices.slice(0, limit) : devices;
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[650px] text-left">
        <thead>
          <tr className="border-b border-border text-[10px] uppercase tracking-[.12em] text-muted-foreground">
            <th className="pb-3 font-semibold">Device</th>
            <th className="pb-3 font-semibold">Type</th>
            <th className="pb-3 font-semibold">Status</th>
            <th className="pb-3 font-semibold">Site</th>
            <th className="pb-3 text-right font-semibold">Health</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((device) => (
            <tr
              key={device.id}
              className="border-b border-border/70 last:border-0"
            >
              <td className="py-3.5">
                <div className="flex items-center gap-3">
                  <div className="grid size-8 place-items-center rounded-lg bg-slate-100 text-slate-600">
                    {device.type === "PCS" ? (
                      <Gauge className="size-4" />
                    ) : device.type === "BMS" ? (
                      <BatteryCharging className="size-4" />
                    ) : (
                      <Network className="size-4" />
                    )}
                  </div>
                  <span className="text-sm font-medium">
                    {device.externalId}
                  </span>
                </div>
              </td>
              <td className="py-3.5 text-xs text-muted-foreground">
                {device.type.replace("_", " ")}
              </td>
              <td className="py-3.5">
                <Badge
                  className={cn(
                    "w-fit",
                    device.status === "ONLINE"
                      ? "border-emerald-100 bg-emerald-50 text-emerald-700"
                      : "border-red-100 bg-red-50 text-red-700",
                  )}
                >
                  {device.status}
                </Badge>
              </td>
              <td className="py-3.5 text-xs text-muted-foreground">
                {siteCode}
              </td>
              <td className="py-3.5 text-right text-xs font-medium text-emerald-700">
                Healthy
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export function OverviewPage({
  data,
  refresh,
  selectSite,
  selectedSite,
}: {
  data: DashboardData;
  refresh: () => void;
  selectSite: (id: string) => void;
  selectedSite?: string;
}) {
  const chart = useEnergyChart(data.history);
  const t = data.telemetry;
  const online = data.devices.filter((d) => d.status === "ONLINE").length;
  const activeAlerts = data.alerts.filter((a) => a.active);
  return (
    <>
      <PageHeader
        page="overview"
        actions={
          <>
            <Select value={selectedSite} onValueChange={selectSite}>
              <SelectTrigger
                aria-label="Select site"
                className="w-full sm:w-auto"
              >
                <SelectValue placeholder="Choose a site" />
              </SelectTrigger>
              <SelectContent>
                {data.sites.map((site) => (
                  <SelectItem key={site.id} value={site.id}>
                    {site.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button variant="outline" onClick={refresh}>
              <Activity className="size-4" />
              Refresh
            </Button>
          </>
        }
      />
      <div className="mt-7 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <MetricCard
          label="Solar production"
          value={formatPower(t?.avgSolarW).split(" ")[0]}
          unit={formatPower(t?.avgSolarW).split(" ")[1] ?? "W"}
          delta="+12.4%"
          icon={Sun}
          tone="bg-amber-50 text-amber-600"
        />
        <MetricCard
          label="Site consumption"
          value={formatPower(t?.avgLoadW).split(" ")[0]}
          unit={formatPower(t?.avgLoadW).split(" ")[1] ?? "W"}
          delta="Stable"
          icon={Building2}
          tone="bg-violet-50 text-violet-600"
        />
        <MetricCard
          label="Battery SOC"
          value={(t?.avgSoc ?? 0).toFixed(1)}
          unit="%"
          delta="Optimal"
          icon={BatteryCharging}
          tone="bg-emerald-50 text-emerald-700"
        />
        <MetricCard
          label="Grid exchange"
          value={formatPower(t?.avgGridW).split(" ")[0]}
          unit={formatPower(t?.avgGridW).split(" ")[1] ?? "W"}
          delta="Live"
          icon={Zap}
          tone="bg-cyan-50 text-cyan-700"
        />
      </div>
      <div className="mt-4 grid gap-4 xl:grid-cols-[1.55fr_1fr]">
        <Card>
          <CardHeader className="flex-row items-start justify-between">
            <div>
              <CardTitle>Energy performance</CardTitle>
              <CardDescription>
                Live power flow over the selected period
              </CardDescription>
            </div>
            <Button variant="ghost" size="icon">
              <MoreHorizontal className="size-5" />
            </Button>
          </CardHeader>
          <CardContent>
            <ReactEChartsCore
              echarts={echarts}
              option={chart}
              style={{ height: 320 }}
            />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <div className="flex items-start justify-between">
              <div>
                <CardTitle>Live energy flow</CardTitle>
                <CardDescription>{data.site?.name}</CardDescription>
              </div>
              <Badge
                className={cn(
                  data.site?.status === "ONLINE"
                    ? "border-emerald-200 bg-emerald-50 text-emerald-700"
                    : "border-red-200 bg-red-50 text-red-700",
                )}
              >
                {data.site?.status ?? "UNKNOWN"}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <EnergyFlow telemetry={t} />
            <div className="mt-2 flex justify-between border-t border-border pt-4 text-xs text-muted-foreground">
              <span>
                Updated{" "}
                {data.updatedAt
                  ? new Date(data.updatedAt).toLocaleTimeString()
                  : "—"}
              </span>
              <span className="font-medium text-foreground">
                SOC {t?.avgSoc?.toFixed(1) ?? "0.0"}%
              </span>
            </div>
          </CardContent>
        </Card>
      </div>
      <div className="mt-4 grid gap-4 xl:grid-cols-[1.55fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Connected assets</CardTitle>
            <CardDescription>
              {online} of {data.devices.length} devices online
            </CardDescription>
          </CardHeader>
          <CardContent>
            <DeviceTable
              devices={data.devices}
              siteCode={data.site?.siteCode}
              limit={6}
            />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Operational status</CardTitle>
            <CardDescription>Live system diagnostics</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {[
              {
                label: "Cloud ingestion",
                detail: "Receiving telemetry",
                icon: Cloud,
                status: "Healthy",
              },
              {
                label: "MQTT broker",
                detail: "QoS 1 · Connected",
                icon: Network,
                status: "Healthy",
              },
              {
                label: "Control engine",
                detail: "1 second cycle",
                icon: Gauge,
                status: "Active",
              },
              {
                label: "Active alerts",
                detail: activeAlerts.length
                  ? `${activeAlerts.length} require review`
                  : "No critical alerts",
                icon: AlertTriangle,
                status: activeAlerts.length ? "Review" : "Clear",
              },
            ].map(({ label, detail, icon: Icon, status }) => (
              <div
                key={label}
                className="flex items-center gap-3 rounded-xl border border-border/70 p-3"
              >
                <div className="grid size-9 place-items-center rounded-lg bg-slate-100 text-slate-600">
                  <Icon className="size-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-sm font-medium">{label}</div>
                  <div className="text-[11px] text-muted-foreground">
                    {detail}
                  </div>
                </div>
                <span
                  className={cn(
                    "text-[11px] font-semibold",
                    status === "Review" ? "text-amber-600" : "text-emerald-700",
                  )}
                >
                  {status}
                </span>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </>
  );
}

export function EnergyPage({ data }: { data: DashboardData }) {
  const chart = useEnergyChart(data.history);
  return (
    <>
      <PageHeader
        page="energy"
        actions={
          <Button variant="outline">
            <CalendarDays className="size-4" />
            Last 24 hours
          </Button>
        }
      />
      <div className="mt-7 grid gap-4 xl:grid-cols-[1.4fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Power flow timeline</CardTitle>
            <CardDescription>
              Near real-time readings across all energy channels
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ReactEChartsCore
              echarts={echarts}
              option={chart}
              style={{ height: 390 }}
            />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Flow right now</CardTitle>
            <CardDescription>{data.site?.name}</CardDescription>
          </CardHeader>
          <CardContent>
            <EnergyFlow telemetry={data.telemetry} />
          </CardContent>
        </Card>
      </div>
      <div className="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {[
          {
            label: "Solar yield",
            value: "184.2 kWh",
            icon: Sun,
            tone: "text-amber-600 bg-amber-50",
          },
          {
            label: "Load served",
            value: "221.8 kWh",
            icon: Building2,
            tone: "text-violet-600 bg-violet-50",
          },
          {
            label: "Battery cycles",
            value: "1.42",
            icon: BatteryCharging,
            tone: "text-emerald-700 bg-emerald-50",
          },
          {
            label: "Grid import",
            value: "37.6 kWh",
            icon: Zap,
            tone: "text-cyan-700 bg-cyan-50",
          },
        ].map(({ label, value, icon: Icon, tone }) => (
          <Card key={label}>
            <CardContent className="flex items-center gap-4 p-5">
              <div
                className={cn(
                  "grid size-11 place-items-center rounded-xl",
                  tone,
                )}
              >
                <Icon className="size-5" />
              </div>
              <div>
                <div className="text-xs text-muted-foreground">{label}</div>
                <div className="mt-1 font-display text-xl font-semibold">
                  {value}
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </>
  );
}

export function SitesPage({
  data,
  tenant,
}: {
  data: DashboardData;
  tenant: TenantKey;
}) {
  const demoSites =
    tenant === "tenant-b"
      ? [
          {
            name: "Osaka Solar Hub",
            code: "jp-osaka-02",
            location: "Osaka, Japan",
            status: "ONLINE",
            capacity: "2.8 MW",
          },
          {
            name: "Kobe Battery Park",
            code: "jp-kobe-01",
            location: "Kobe, Japan",
            status: "MAINTENANCE",
            capacity: "4.1 MWh",
          },
        ]
      : data.sites.map((s) => ({
          name: s.name,
          code: s.siteCode,
          location: s.location ?? "Singapore",
          status: s.status,
          capacity: "3.6 MWh",
        }));
  return (
    <>
      <PageHeader
        page="sites"
        actions={
          <Button>
            <Plus className="size-4" />
            Add site
          </Button>
        }
      />
      {tenant === "tenant-b" && (
        <div className="mt-6 flex items-start gap-3 rounded-xl border border-violet-200 bg-violet-50 p-4 text-sm text-violet-800">
          <Radio className="mt-0.5 size-4 shrink-0" />
          <div>
            <span className="font-semibold">Preview dataset.</span> Tenant B
            demonstrates multi-tenant portfolio UI and is not connected to the
            live simulator.
          </div>
        </div>
      )}
      <div className="mt-6 grid gap-4 md:grid-cols-2 2xl:grid-cols-3">
        {demoSites.map((site, index) => (
          <Card key={site.code} className="overflow-hidden">
            <div
              className={cn(
                "h-1.5",
                index % 2 ? "bg-violet-500" : "bg-emerald-500",
              )}
            />
            <CardContent className="p-5">
              <div className="flex items-start justify-between">
                <div className="grid size-11 place-items-center rounded-xl bg-slate-100 text-slate-700">
                  <Building2 className="size-5" />
                </div>
                <Badge
                  className={cn(
                    site.status === "ONLINE"
                      ? "border-emerald-100 bg-emerald-50 text-emerald-700"
                      : "border-amber-100 bg-amber-50 text-amber-700",
                  )}
                >
                  {site.status}
                </Badge>
              </div>
              <h3 className="mt-5 font-display text-lg font-semibold">
                {site.name}
              </h3>
              <div className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
                <MapPin className="size-3.5" />
                {site.location}
              </div>
              <div className="mt-5 grid grid-cols-2 gap-3 rounded-xl bg-slate-50 p-3">
                <div>
                  <div className="text-[10px] uppercase tracking-wider text-muted-foreground">
                    Site code
                  </div>
                  <div className="mt-1 text-xs font-semibold">{site.code}</div>
                </div>
                <div>
                  <div className="text-[10px] uppercase tracking-wider text-muted-foreground">
                    Capacity
                  </div>
                  <div className="mt-1 text-xs font-semibold">
                    {site.capacity}
                  </div>
                </div>
              </div>
              <Button variant="outline" className="mt-4 w-full">
                Open site
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
    </>
  );
}

export function DevicesPage({ data }: { data: DashboardData }) {
  return (
    <>
      <PageHeader
        page="devices"
        actions={
          <Button>
            <Plus className="size-4" />
            Provision device
          </Button>
        }
      />
      <Card className="mt-7">
        <CardHeader className="gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <CardTitle>Device inventory</CardTitle>
            <CardDescription>
              {data.devices.length} registered assets
            </CardDescription>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-400" />
              <Input
                className="h-10 bg-white pl-9 sm:w-64"
                placeholder="Search device…"
              />
            </div>
            <Button variant="outline">
              <Filter className="size-4" />
              Filter
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <DeviceTable devices={data.devices} siteCode={data.site?.siteCode} />
        </CardContent>
      </Card>
    </>
  );
}

export function CommandsPage() {
  const rows = [
    {
      name: "Set PCS active power",
      target: "pcs-01",
      status: "ACKNOWLEDGED",
      time: "10:24:08",
    },
    {
      name: "Update charge limit",
      target: "bms-02",
      status: "COMPLETED",
      time: "09:48:42",
    },
    {
      name: "Read inverter state",
      target: "pv-inverter-01",
      status: "COMPLETED",
      time: "09:31:17",
    },
  ];
  return (
    <>
      <PageHeader
        page="commands"
        actions={
          <Button>
            <Play className="size-4" />
            New command
          </Button>
        }
      />
      <div className="mt-7 grid gap-4 xl:grid-cols-[1fr_1.7fr]">
        <Card>
          <CardHeader>
            <CardTitle>Quick actions</CardTitle>
            <CardDescription>
              Commonly used, policy-safe commands
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {[
              "Set ESS charge limit",
              "Start demand response",
              "Request device status",
            ].map((item, index) => (
              <button
                key={item}
                className="flex w-full items-center gap-3 rounded-xl border border-border p-4 text-left transition hover:border-emerald-200 hover:bg-emerald-50/50"
              >
                <div className="grid size-9 place-items-center rounded-lg bg-emerald-50 text-emerald-700">
                  {index === 0 ? (
                    <BatteryCharging className="size-4" />
                  ) : index === 1 ? (
                    <Activity className="size-4" />
                  ) : (
                    <Radio className="size-4" />
                  )}
                </div>
                <span className="flex-1 text-sm font-medium">{item}</span>
                <ArrowUpRight className="size-4 text-slate-400" />
              </button>
            ))}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Recent command log</CardTitle>
            <CardDescription>
              Immutable audit trail for remote actions
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {rows.map((row) => (
              <div
                key={row.time}
                className="flex flex-col gap-3 rounded-xl border border-border/70 p-4 sm:flex-row sm:items-center"
              >
                <div className="grid size-9 place-items-center rounded-lg bg-slate-100">
                  <Command className="size-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-medium">{row.name}</div>
                  <div className="text-xs text-muted-foreground">
                    {row.target} · {row.time}
                  </div>
                </div>
                <Badge className="w-fit border-emerald-100 bg-emerald-50 text-emerald-700">
                  {row.status}
                </Badge>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </>
  );
}

export function AlertsPage({ data }: { data: DashboardData }) {
  const demo = data.alerts.length
    ? data.alerts
    : [
        {
          id: "a1",
          severity: "WARNING",
          message: "Battery temperature approaching configured threshold",
          active: true,
          createdAt: new Date().toISOString(),
        },
        {
          id: "a2",
          severity: "INFO",
          message: "Scheduled maintenance window starts in 2 days",
          active: false,
          createdAt: new Date().toISOString(),
        },
      ];
  return (
    <>
      <PageHeader
        page="alerts"
        actions={
          <Button variant="outline">
            <Settings2 className="size-4" />
            Alert rules
          </Button>
        }
      />
      <div className="mt-7 grid gap-4 lg:grid-cols-3">
        <MetricCard
          label="Active alerts"
          value={String(demo.filter((a) => a.active).length)}
          unit="open"
          delta="Live"
          icon={AlertTriangle}
          tone="bg-amber-50 text-amber-600"
        />
        <MetricCard
          label="Acknowledged"
          value="12"
          unit="today"
          delta="On track"
          icon={CheckCircle2}
          tone="bg-emerald-50 text-emerald-700"
        />
        <MetricCard
          label="Mean response"
          value="4.2"
          unit="min"
          delta="-18%"
          icon={Activity}
          tone="bg-cyan-50 text-cyan-700"
        />
      </div>
      <Card className="mt-4">
        <CardHeader>
          <CardTitle>Alert feed</CardTitle>
          <CardDescription>Newest and highest priority first</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {demo.map((alert) => (
            <div
              key={alert.id}
              className="flex flex-col gap-3 rounded-xl border border-border p-4 sm:flex-row sm:items-center"
            >
              <div
                className={cn(
                  "grid size-10 place-items-center rounded-xl",
                  alert.active
                    ? "bg-amber-50 text-amber-700"
                    : "bg-slate-100 text-slate-500",
                )}
              >
                <AlertTriangle className="size-5" />
              </div>
              <div className="min-w-0 flex-1">
                <div className="text-sm font-medium">{alert.message}</div>
                <div className="mt-1 text-xs text-muted-foreground">
                  {new Date(alert.createdAt).toLocaleString()} ·{" "}
                  {alert.severity}
                </div>
              </div>
              {alert.active ? (
                <Button size="sm" variant="outline">
                  Acknowledge
                </Button>
              ) : (
                <Badge>Resolved</Badge>
              )}
            </div>
          ))}
        </CardContent>
      </Card>
    </>
  );
}

export function ReportsPage() {
  const option = {
    color: ["#0c7664", "#d9a01b"],
    tooltip: { trigger: "axis" },
    grid: { left: 8, right: 8, top: 20, bottom: 4, containLabel: true },
    xAxis: {
      type: "category",
      data: ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"],
      axisTick: { show: false },
      axisLine: { show: false },
    },
    yAxis: { type: "value", splitLine: { lineStyle: { color: "#eef0eb" } } },
    series: [
      {
        type: "bar",
        name: "Consumption",
        data: [172, 194, 188, 221, 209, 181, 167],
        barWidth: 12,
        itemStyle: { borderRadius: [6, 6, 0, 0] },
      },
      {
        type: "bar",
        name: "Solar",
        data: [114, 121, 143, 152, 149, 134, 128],
        barWidth: 12,
        itemStyle: { borderRadius: [6, 6, 0, 0] },
      },
    ],
  };
  return (
    <>
      <PageHeader
        page="reports"
        actions={
          <>
            <Button variant="outline">
              <CalendarDays className="size-4" />
              This week
            </Button>
            <Button>
              <Download className="size-4" />
              Export PDF
            </Button>
          </>
        }
      />
      <div className="mt-7 grid gap-4 xl:grid-cols-[1.6fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Energy summary</CardTitle>
            <CardDescription>Daily generation and consumption</CardDescription>
          </CardHeader>
          <CardContent>
            <ReactEChartsCore
              echarts={echarts}
              option={option}
              style={{ height: 360 }}
            />
          </CardContent>
        </Card>
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-1">
          <Card>
            <CardContent className="p-5">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">
                    Energy cost avoided
                  </p>
                  <div className="mt-2 font-display text-3xl font-semibold">
                    $2,840
                  </div>
                </div>
                <div className="grid size-11 place-items-center rounded-xl bg-emerald-50 text-emerald-700">
                  <ArrowDownRight className="size-5" />
                </div>
              </div>
              <p className="mt-3 text-xs text-emerald-700">
                12.8% better than last week
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-5">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">
                    Carbon avoided
                  </p>
                  <div className="mt-2 font-display text-3xl font-semibold">
                    1.24 t
                  </div>
                </div>
                <div className="grid size-11 place-items-center rounded-xl bg-cyan-50 text-cyan-700">
                  <Leaf className="size-5" />
                </div>
              </div>
              <p className="mt-3 text-xs text-muted-foreground">
                Equivalent to 57 mature trees
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}

export function SettingsPage({ tenant }: { tenant: TenantKey }) {
  return (
    <>
      <PageHeader page="settings" />
      <div className="mt-7 grid gap-4 xl:grid-cols-[1.5fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Workspace profile</CardTitle>
            <CardDescription>
              Defaults shared across your tenant
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-5 sm:grid-cols-2">
            <label className="space-y-2 text-sm font-medium">
              Workspace name
              <Input
                defaultValue={
                  tenant === "tenant-a" ? "Demo Tenant A" : "Demo Tenant B"
                }
              />
            </label>
            <label className="space-y-2 text-sm font-medium">
              Timezone
              <Select
                defaultValue={
                  tenant === "tenant-a" ? "asia-singapore" : "asia-tokyo"
                }
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="asia-singapore">
                    Asia/Singapore (UTC+8)
                  </SelectItem>
                  <SelectItem value="asia-tokyo">Asia/Tokyo (UTC+9)</SelectItem>
                  <SelectItem value="asia-saigon">
                    Asia/Ho Chi Minh (UTC+7)
                  </SelectItem>
                </SelectContent>
              </Select>
            </label>
            <label className="space-y-2 text-sm font-medium">
              Energy unit
              <Select defaultValue="kwh">
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="kwh">Kilowatt-hour (kWh)</SelectItem>
                  <SelectItem value="mwh">Megawatt-hour (MWh)</SelectItem>
                </SelectContent>
              </Select>
            </label>
            <label className="space-y-2 text-sm font-medium">
              Currency
              <Select defaultValue="usd">
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="usd">USD ($)</SelectItem>
                  <SelectItem value="sgd">SGD (S$)</SelectItem>
                  <SelectItem value="jpy">JPY (¥)</SelectItem>
                </SelectContent>
              </Select>
            </label>
            <div className="sm:col-span-2">
              <Button>Save changes</Button>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Security</CardTitle>
            <CardDescription>Tenant access and authentication</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {[
              {
                label: "Multi-factor authentication",
                detail: "Required for administrators",
                icon: ShieldCheck,
              },
              {
                label: "Workspace members",
                detail: "8 active users",
                icon: Users,
              },
              {
                label: "Audit logging",
                detail: "365-day retention",
                icon: Activity,
              },
            ].map(({ label, detail, icon: Icon }) => (
              <div
                key={label}
                className="flex items-center gap-3 rounded-xl border border-border p-3"
              >
                <div className="grid size-9 place-items-center rounded-lg bg-emerald-50 text-emerald-700">
                  <Icon className="size-4" />
                </div>
                <div className="flex-1">
                  <div className="text-sm font-medium">{label}</div>
                  <div className="text-xs text-muted-foreground">{detail}</div>
                </div>
                <CheckCircle2 className="size-4 text-emerald-600" />
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </>
  );
}

export function ProfilePage() {
  return (
    <>
      <PageHeader page="profile" actions={<Button>Save profile</Button>} />
      <div className="mt-7 grid gap-4 xl:grid-cols-[1.45fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Personal information</CardTitle>
            <CardDescription>
              Shown across your workspace and audit records
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="mb-6 flex flex-col gap-4 border-b border-border pb-6 sm:flex-row sm:items-center">
              <div className="grid size-20 place-items-center rounded-2xl bg-gradient-to-br from-emerald-700 to-emerald-500 font-display text-2xl font-semibold text-white shadow-lg">
                DA
              </div>
              <div className="flex-1">
                <div className="font-display text-lg font-semibold">
                  Demo Admin
                </div>
                <div className="mt-1 text-sm text-muted-foreground">
                  Administrator · CDI Energy Cloud
                </div>
                <Button variant="outline" size="sm" className="mt-3">
                  Change avatar
                </Button>
              </div>
            </div>
            <div className="grid gap-5 sm:grid-cols-2">
              <label className="space-y-2 text-sm font-medium">
                First name
                <Input defaultValue="Demo" />
              </label>
              <label className="space-y-2 text-sm font-medium">
                Last name
                <Input defaultValue="Admin" />
              </label>
              <label className="space-y-2 text-sm font-medium sm:col-span-2">
                Email address
                <Input type="email" defaultValue="admin@demo.local" />
              </label>
              <label className="space-y-2 text-sm font-medium">
                Job title
                <Input defaultValue="Energy Operations Manager" />
              </label>
              <label className="space-y-2 text-sm font-medium">
                Language
                <Select defaultValue="en">
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="en">English</SelectItem>
                    <SelectItem value="vi">Tiếng Việt</SelectItem>
                    <SelectItem value="ja">日本語</SelectItem>
                  </SelectContent>
                </Select>
              </label>
            </div>
          </CardContent>
        </Card>
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Account security</CardTitle>
              <CardDescription>Keep your access protected</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center gap-3 rounded-xl border border-border p-4">
                <div className="grid size-10 place-items-center rounded-xl bg-emerald-50 text-emerald-700">
                  <KeyRound className="size-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-sm font-medium">Password</div>
                  <div className="text-xs text-muted-foreground">
                    Updated 24 days ago
                  </div>
                </div>
                <Button variant="outline" size="sm">
                  Change
                </Button>
              </div>
              <div className="flex items-center gap-3 rounded-xl border border-border p-4">
                <div className="grid size-10 place-items-center rounded-xl bg-violet-50 text-violet-700">
                  <ShieldCheck className="size-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-sm font-medium">
                    Two-factor authentication
                  </div>
                  <div className="text-xs text-emerald-700">Enabled</div>
                </div>
                <Button variant="outline" size="sm">
                  Manage
                </Button>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Active session</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-start gap-3">
                <div className="mt-1 size-2 rounded-full bg-emerald-500" />
                <div>
                  <div className="text-sm font-medium">Codex local preview</div>
                  <div className="mt-1 text-xs leading-5 text-muted-foreground">
                    Ho Chi Minh City · Active now
                    <br />
                    Chrome on macOS
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}

export function HelpPage() {
  const topics = [
    {
      title: "Getting started",
      detail: "Workspace setup, users and first site",
      icon: BookOpen,
    },
    {
      title: "Devices & telemetry",
      detail: "Provisioning, connectivity and data quality",
      icon: Network,
    },
    {
      title: "Commands & controls",
      detail: "Safe remote actions and audit trails",
      icon: Command,
    },
    {
      title: "Account & security",
      detail: "Roles, access and authentication",
      icon: ShieldCheck,
    },
  ];
  const questions = [
    "How do I add a new site?",
    "Why is a device showing as offline?",
    "How are commands acknowledged?",
    "How do I export an energy report?",
  ];
  return (
    <>
      <PageHeader page="help" />
      <div className="mt-7 rounded-2xl bg-[#0b1c18] p-5 text-white sm:p-8">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="font-display text-2xl font-semibold sm:text-3xl">
            Search the knowledge base
          </h2>
          <p className="mt-2 text-sm text-slate-400">
            Guides and answers for operating your energy fleet.
          </p>
          <div className="relative mt-6">
            <Search className="absolute left-4 top-1/2 size-5 -translate-y-1/2 text-slate-400" />
            <Input
              className="h-13 border-white/10 bg-white pl-12 text-slate-950 shadow-xl"
              placeholder="Search guides, devices, commands…"
            />
          </div>
        </div>
      </div>
      <div className="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {topics.map(({ title, detail, icon: Icon }) => (
          <Card
            key={title}
            className="group cursor-pointer transition hover:-translate-y-0.5 hover:border-emerald-200 hover:shadow-lg"
          >
            <CardContent className="p-5">
              <div className="flex items-start justify-between">
                <div className="grid size-11 place-items-center rounded-xl bg-emerald-50 text-emerald-700">
                  <Icon className="size-5" />
                </div>
                <ChevronRight className="size-4 text-slate-300 transition group-hover:translate-x-1 group-hover:text-emerald-600" />
              </div>
              <h3 className="mt-5 text-sm font-semibold">{title}</h3>
              <p className="mt-1 text-xs leading-5 text-muted-foreground">
                {detail}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>
      <div className="mt-4 grid gap-4 xl:grid-cols-[1.45fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Frequently asked questions</CardTitle>
            <CardDescription>
              Quick answers from our support team
            </CardDescription>
          </CardHeader>
          <CardContent className="divide-y divide-border">
            {questions.map((question) => (
              <button
                key={question}
                className="flex w-full items-center justify-between gap-4 py-4 text-left text-sm font-medium hover:text-emerald-700"
              >
                {question}
                <ChevronRight className="size-4 shrink-0 text-slate-400" />
              </button>
            ))}
          </CardContent>
        </Card>
        <Card className="overflow-hidden">
          <CardContent className="flex h-full flex-col bg-gradient-to-br from-emerald-50 to-cyan-50 p-6">
            <div className="grid size-12 place-items-center rounded-2xl bg-white text-emerald-700 shadow-sm">
              <MessageCircle className="size-5" />
            </div>
            <h3 className="mt-6 font-display text-xl font-semibold">
              Still need help?
            </h3>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              Our energy operations support team is available around the clock.
            </p>
            <div className="mt-auto space-y-2 pt-8">
              <Button className="w-full">
                <MessageCircle className="size-4" />
                Start support chat
              </Button>
              <Button variant="outline" className="w-full bg-white">
                <Mail className="size-4" />
                Email support
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
