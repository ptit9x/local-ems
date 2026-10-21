import * as React from 'react';
import { cn } from '../../lib/utils';

export function Input({ className, type, ref, ...props }: React.ComponentProps<'input'>) {
  return <input ref={ref} type={type} className={cn('flex h-11 w-full rounded-lg border border-border bg-background/60 px-3 py-2 text-sm outline-none transition placeholder:text-muted-foreground focus:border-primary/70 focus:ring-4 focus:ring-primary/10 disabled:opacity-50', className)} {...props} />;
}
