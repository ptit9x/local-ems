import {
  Activity,
  CheckCircle2,
  ShieldCheck,
  Users,
} from "lucide-react";
import { Button } from "../../../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import { Input } from "../../../components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";
import type { TenantKey } from "../../../dashboard-types";
import { PageHeader } from "../../../components/dashboard/shared";

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
