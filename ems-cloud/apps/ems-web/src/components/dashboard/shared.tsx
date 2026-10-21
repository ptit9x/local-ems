import { useMemo } from "react";
import { EChart } from "../ui/echart";
import * as echarts from "echarts/core";
import { BarChart, LineChart } from "echarts/charts";
import {
  GridComponent,
  LegendComponent,
  TooltipComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import {
  BatteryCharging,
  Building2,
  Cloud,
  Gauge,
  Leaf,
  Network,
  Sun,
  Zap,
} from "lucide-react";
import { Badge } from "../ui/badge";
import {
  Card,
  CardContent,
} from "../ui/card";
import type {
  Device,
  PageKey,
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

export { echarts };

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
    description: "Here's what's happening across your energy fleet today.",
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

export function MetricCard({
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

export function useEnergyChart(history: Telemetry[]) {
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

export function DeviceTable({
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
