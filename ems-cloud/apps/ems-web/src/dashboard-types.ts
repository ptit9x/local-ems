export type PageKey =
  | "overview"
  | "energy"
  | "sites"
  | "devices"
  | "commands"
  | "alerts"
  | "reports"
  | "settings"
  | "profile"
  | "help";
export type TenantKey = "tenant-a" | "tenant-b";
export type Site = {
  id: string;
  siteCode: string;
  name: string;
  status: string;
  lastSeenAt?: string;
  location?: string;
};
export type Telemetry = {
  timestamp: string;
  avgSolarW: number;
  avgLoadW: number;
  avgGridW: number;
  avgEssW: number;
  avgSoc: number;
  maxTempC?: number;
};
export type Device = {
  id: string;
  externalId: string;
  type: string;
  status: string;
};
export type Alert = {
  id: string;
  severity: string;
  message: string;
  active: boolean;
  createdAt: string;
};
export type DashboardData = {
  site?: Site;
  sites: Site[];
  telemetry?: Telemetry;
  history: Telemetry[];
  devices: Device[];
  alerts: Alert[];
  updatedAt: number;
};
