import {
  Activity,
  AlertTriangle,
  BatteryCharging,
  Building2,
  Cloud,
  Gauge,
  MoreHorizontal,
  Network,
  Sun,
  Zap,
} from "lucide-react";
import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";
import type { DashboardData } from "../../../dashboard-types";
import { cn } from "../../../lib/utils";
import {
  echarts,
  DeviceTable,
  EnergyFlow,
  formatPower,
  MetricCard,
  PageHeader,
  useEnergyChart,
} from "../../../components/dashboard/shared";
import { EChart } from "../../../components/ui/echart";

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
            <EChart
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
