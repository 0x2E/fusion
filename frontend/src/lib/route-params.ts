export function parsePositiveIntegerParam(value: unknown): number | null {
  if (typeof value === "number") {
    if (!Number.isSafeInteger(value) || value <= 0) {
      return null;
    }

    return value;
  }

  if (typeof value !== "string") {
    return null;
  }

  if (!/^[1-9]\d*$/.test(value)) {
    return null;
  }

  const parsed = Number.parseInt(value, 10);
  return Number.isSafeInteger(parsed) ? parsed : null;
}
