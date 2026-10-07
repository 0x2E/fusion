// Matches the backend whitelist in handler/feed.go; GET /api/app returns the
// authoritative list and this is only the pre-load fallback.
export const FALLBACK_REFRESH_INTERVALS = [
  900, 1800, 3600, 7200, 21600, 43200, 86400,
] as const;

export const DEFAULT_PULL_INTERVAL = 1800;

// Formats a duration in seconds as "15m" / "1h" / "24h". Values that are not
// whole hours fall back to rounded minutes so arbitrary global intervals
// never render as fractional hours. Locale-neutral by design: no i18n keys.
export function formatInterval(seconds: number): string {
  if (seconds >= 86400 && seconds % 86400 === 0) return `${seconds / 86400}d`;
  if (seconds >= 3600 && seconds % 3600 === 0) return `${seconds / 3600}h`;
  return `${Math.round(seconds / 60)}m`;
}

// The dropdown options: the allowed presets plus the global interval, sorted.
// The entry equal to the global interval means "follow the global setting"
// (stored as NULL); every other entry is an explicit override.
export function intervalOptions(
  globalInterval: number,
  allowed: readonly number[],
): number[] {
  const values = allowed.includes(globalInterval)
    ? [...allowed]
    : [...allowed, globalInterval];
  return values.sort((a, b) => a - b);
}
