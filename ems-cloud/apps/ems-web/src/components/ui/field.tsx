import * as React from 'react';
import { cn } from '../../lib/utils';

export function Field({ className, ...props }: React.ComponentProps<'div'>) {
  return <div className={cn('space-y-2', className)} {...props} />;
}

export function FieldLabel({ className, ...props }: React.ComponentProps<'label'>) {
  return <label className={cn('block text-sm font-medium text-foreground', className)} {...props} />;
}

export function FieldDescription({ className, ...props }: React.ComponentProps<'p'>) {
  return <p className={cn('text-xs leading-5 text-muted-foreground', className)} {...props} />;
}

export function FieldError({ className, ...props }: React.ComponentProps<'p'>) {
  return <p role="alert" className={cn('text-xs font-medium text-destructive', className)} {...props} />;
}
