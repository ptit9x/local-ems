import {
  Building2,
  MapPin,
  Plus,
  Radio,
} from "lucide-react";
import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import {
  Card,
  CardContent,
} from "../../../components/ui/card";
import type { DashboardData, TenantKey } from "../../../dashboard-types";
import { cn } from "../../../lib/utils";
import { PageHeader } from "../../../components/dashboard/shared";

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
