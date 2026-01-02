import { AlertCircleIcon } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface ErrorStateProps {
  title?: string;
  message: string;
  onRetry?: () => void;
  retrying?: boolean;
  variant?: 'inline' | 'centered';
  className?: string;
}

export function ErrorState({
  title = 'Failed to load',
  message,
  onRetry,
  retrying = false,
  variant = 'centered',
  className,
}: ErrorStateProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center gap-3',
        variant === 'centered' ? 'py-12 px-4' : 'py-6 px-3',
        className
      )}
    >
      <div className="rounded-md bg-destructive/10 p-4 text-center">
        <div className="flex items-center justify-center gap-2 mb-1">
          <AlertCircleIcon className="size-4 text-destructive" />
          <p className="text-sm font-medium text-destructive">{title}</p>
        </div>
        <p className="text-xs text-destructive/80 mt-1">{message}</p>
      </div>
      {onRetry && (
        <Button
          variant="outline"
          size="sm"
          onClick={onRetry}
          disabled={retrying}
        >
          {retrying ? (
            <>
              <div className="size-3 mr-2 animate-spin rounded-full border-2 border-current border-t-transparent" />
              Retrying...
            </>
          ) : (
            'Retry'
          )}
        </Button>
      )}
    </div>
  );
}
