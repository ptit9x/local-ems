import {
  BatteryCharging,
  Building2,
  CalendarDays,
  Sun,
  Zap,
} from "lucide-react";
import { Button } from "../../../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import type { DashboardData } from "../../../dashboard-types";
import { cn } from "../../../lib/utils";
import { echarts, EnergyFlow, PageHeader, useEnergyChart } from "../../../components/dashboard/shared";
import { EChart } from "../../../components/ui/echart";

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
            <EChart
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
