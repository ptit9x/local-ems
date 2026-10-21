import { useRef, useEffect } from "react";
import type { ECharts, EChartsCoreOption } from "echarts/core";

interface EChartProps {
  echarts: typeof import("echarts/core");
  option: EChartsCoreOption;
  style?: React.CSSProperties;
}

export function EChart({ echarts: ec, option, style }: EChartProps) {
  const ref = useRef<HTMLDivElement>(null);
  const chartRef = useRef<ECharts | null>(null);

  useEffect(() => {
    if (!ref.current) return;
    if (!chartRef.current) {
      chartRef.current = ec.init(ref.current);
    }
    chartRef.current.setOption(option, true);
  }, [ec, option]);

  useEffect(() => {
    const onResize = () => chartRef.current?.resize();
    window.addEventListener("resize", onResize);
    return () => {
      window.removeEventListener("resize", onResize);
      chartRef.current?.dispose();
      chartRef.current = null;
    };
  }, []);

  return <div ref={ref} style={style} />;
}
