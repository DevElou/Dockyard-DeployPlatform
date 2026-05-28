import { LoadingState } from "./loading-state";
import { ErrorState } from "./error-state";

interface DataGuardProps<T extends { items: unknown[] }> {
  isLoading: boolean;
  isError: boolean;
  data: T | undefined;
  errorMessage: string;
  onRetry?: () => void;
  loadingRows?: number;
  empty: React.ReactNode;
  children: (data: T) => React.ReactNode;
}

/**
 * Handles the loading / error / empty / ready tristate that every list page needs.
 *
 * Usage:
 *   <DataGuard {...query} errorMessage="Could not load releases." empty={<EmptyState .../>}>
 *     {(data) => <ReleasesTable items={data.items} />}
 *   </DataGuard>
 */
export function DataGuard<T extends { items: unknown[] }>({
  isLoading,
  isError,
  data,
  errorMessage,
  onRetry,
  loadingRows,
  empty,
  children,
}: DataGuardProps<T>) {
  if (isLoading) return <LoadingState rows={loadingRows} />;
  if (isError) return <ErrorState message={errorMessage} onRetry={onRetry} />;
  if (!data || data.items.length === 0) return <>{empty}</>;
  return <>{children(data)}</>;
}
