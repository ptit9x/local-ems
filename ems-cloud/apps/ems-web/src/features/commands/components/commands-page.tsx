import {
  Activity,
  ArrowUpRight,
  BatteryCharging,
  Command,
  Play,
  Radio,
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
import { PageHeader } from "../../../components/dashboard/shared";

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
