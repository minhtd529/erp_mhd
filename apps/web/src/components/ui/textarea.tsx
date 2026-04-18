import * as React from 'react';
import { cn } from '@/lib/utils';

export interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(({ className, ...props }, ref) => (
  <textarea
    className={cn(
      'flex min-h-[80px] w-full rounded border border-border bg-surface px-3 py-2 text-sm text-text-primary shadow-sm',
      'placeholder:text-text-secondary',
      'focus:outline-none focus:border-action focus:ring-1 focus:ring-action',
      'disabled:cursor-not-allowed disabled:opacity-50 resize-none',
      className
    )}
    ref={ref}
    {...props}
  />
));
Textarea.displayName = 'Textarea';

export { Textarea };
