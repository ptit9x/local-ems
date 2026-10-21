import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  Settings2,
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
import type { DashboardData } from "../../../dashboard-types";
import { cn } from "../../../lib/utils";
import { MetricCard, PageHeader } from "../../../components/dashboard/shared";

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
