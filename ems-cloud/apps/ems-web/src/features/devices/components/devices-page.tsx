import {
  Filter,
  Plus,
  Search,
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
import type { DashboardData } from "../../../dashboard-types";
import { DeviceTable, PageHeader } from "../../../components/dashboard/shared";

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
