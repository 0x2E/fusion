const FEED_ERROR_PREVIEW_LENGTH = 48;

function normalizeFeedError(error: string): string {
  return error.replace(/\s+/g, " ").trim();
}

export function getFeedErrorPreview(error: string): string {
  const normalizedError = normalizeFeedError(error);
  if (normalizedError.length <= FEED_ERROR_PREVIEW_LENGTH) {
    return normalizedError;
  }

  return `${normalizedError.slice(0, FEED_ERROR_PREVIEW_LENGTH)}...`;
}
