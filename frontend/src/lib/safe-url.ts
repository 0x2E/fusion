const ALLOWED_EXTERNAL_PROTOCOLS = new Set(["http:", "https:"]);

function parseURL(raw: string, base?: URL): URL | null {
  const trimmed = raw.trim();
  if (!trimmed) {
    return null;
  }

  try {
    const parsed = base ? new URL(trimmed, base) : new URL(trimmed);
    if (!ALLOWED_EXTERNAL_PROTOCOLS.has(parsed.protocol)) {
      return null;
    }
    if (!parsed.hostname) {
      return null;
    }

    return parsed;
  } catch {
    return null;
  }
}

export function toSafeExternalUrl(raw: string | null | undefined): string | null {
  if (!raw) {
    return null;
  }

  return parseURL(raw)?.href ?? null;
}

export function resolveSafeExternalUrl(
  raw: string | null | undefined,
  baseUrl: string | null | undefined,
): string | null {
  if (!raw) {
    return null;
  }

  const safeBase = baseUrl ? parseURL(baseUrl) : null;
  return parseURL(raw, safeBase ?? undefined)?.href ?? null;
}
