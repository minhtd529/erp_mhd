'use client';
import * as React from 'react';
import { Search } from 'lucide-react';
import { cn } from '@/lib/utils';

interface SearchInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  onSearch?: (value: string) => void;
}

export function SearchInput({ className, onSearch, onChange, ...props }: SearchInputProps) {
  const [value, setValue] = React.useState('');

  React.useEffect(() => {
    const timer = setTimeout(() => onSearch?.(value), 300);
    return () => clearTimeout(timer);
  }, [value, onSearch]);

  return (
    <div className="relative">
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-secondary pointer-events-none" />
      <input
        value={value}
        onChange={(e) => {
          setValue(e.target.value);
          onChange?.(e);
        }}
        className={cn(
          'flex h-9 w-full rounded border border-border bg-surface pl-9 pr-3 py-1 text-sm text-text-primary',
          'placeholder:text-text-secondary',
          'focus:outline-none focus:border-action focus:ring-1 focus:ring-action',
          className
        )}
        {...props}
      />
    </div>
  );
}
