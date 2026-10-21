import { Bolt } from 'lucide-react';

export function Brand() {
  return <div className="flex items-center gap-3"><div className="grid size-10 place-items-center rounded-xl bg-primary text-primary-foreground shadow-glow"><Bolt className="size-5" fill="currentColor" /></div><div><div className="font-display text-base font-semibold tracking-tight">CDI Energy</div><div className="text-[11px] font-medium uppercase tracking-[.18em] text-muted-foreground">Control cloud</div></div></div>;
}
