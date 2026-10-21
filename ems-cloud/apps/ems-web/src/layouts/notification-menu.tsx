import { useState } from "react";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { Bell, Check } from "lucide-react";
import type { Alert } from "../dashboard-types";
import { cn } from "../lib/utils";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";

export function NotificationMenu({
  alerts,
  onViewAll,
}: {
  alerts: Alert[];
  onViewAll: () => void;
}) {
  const [read, setRead] = useState(false);
  const active = alerts.filter((item) => item.active);
  const items = active.length
    ? active.slice(0, 3)
    : [
        {
          id: "system-ready",
          message: "All energy services are operating normally",
          severity: "INFO",
          createdAt: new Date().toISOString(),
          active: false,
        },
      ];
  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild>
        <Button
          aria-label="Notifications"
          variant="ghost"
          size="icon"
          className="relative"
        >
          <Bell className="size-[18px]" />
          {!read && active.length > 0 && (
            <span className="absolute right-2 top-2 size-1.5 rounded-full bg-amber-500 ring-2 ring-[#f4f5f0]" />
          )}
        </Button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          sideOffset={8}
          className="z-50 w-[calc(100vw-2rem)] max-w-[380px] overflow-hidden rounded-2xl border border-border bg-white shadow-2xl"
        >
          <div className="flex items-center justify-between border-b border-border px-4 py-3.5">
            <div>
              <div className="text-sm font-semibold">Notifications</div>
              <div className="text-[11px] text-muted-foreground">
                {active.length} active updates
              </div>
            </div>
            <button
              onClick={() => setRead(true)}
              className="text-xs font-medium text-emerald-700 hover:underline"
            >
              Mark all read
            </button>
          </div>
          <div className="max-h-[340px] overflow-y-auto p-2">
            {items.map((item) => (
              <DropdownMenu.Item
                key={item.id}
                onSelect={onViewAll}
                className="flex cursor-pointer gap-3 rounded-xl p-3 outline-none hover:bg-slate-50 focus:bg-slate-50"
              >
                <div
                  className={cn(
                    "mt-0.5 grid size-9 shrink-0 place-items-center rounded-xl",
                    item.active
                      ? "bg-amber-50 text-amber-700"
                      : "bg-emerald-50 text-emerald-700",
                  )}
                >
                  {item.active ? (
                    <Bell className="size-4" />
                  ) : (
                    <Check className="size-4" />
                  )}
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-sm font-medium leading-5 text-slate-900">
                    {item.message}
                  </div>
                  <div className="mt-1 text-[11px] text-muted-foreground">
                    {item.severity} ·{" "}
                    {new Date(item.createdAt).toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </div>
                </div>
                {!read && item.active && (
                  <span className="mt-2 size-2 shrink-0 rounded-full bg-amber-500" />
                )}
              </DropdownMenu.Item>
            ))}
          </div>
          <div className="border-t border-border p-2">
            <DropdownMenu.Item
              onSelect={onViewAll}
              className="cursor-pointer rounded-lg px-3 py-2.5 text-center text-sm font-medium text-emerald-700 outline-none hover:bg-emerald-50"
            >
              View all notifications
            </DropdownMenu.Item>
          </div>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
