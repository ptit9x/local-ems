import {
  ArrowDownRight,
  CalendarDays,
  Download,
  Leaf,
} from "lucide-react";
import { Button } from "../../../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import { echarts, PageHeader } from "../../../components/dashboard/shared";
import { EChart } from "../../../components/ui/echart";

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
            <EChart
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
