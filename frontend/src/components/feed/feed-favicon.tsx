import { useState } from "react";
import { cn } from "@/lib/utils";

interface FeedFaviconProps {
  src?: string | null;
  className?: string;
}

export function FeedFavicon({ src, className }: FeedFaviconProps) {
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  const loadFailed = !src || failedSrc === src;

  if (!src || loadFailed) {
    return (
      <span
        aria-hidden="true"
        className={cn(
          "inline-block shrink-0 rounded bg-gray-300 dark:bg-gray-600",
          className,
        )}
      />
    );
  }

  return (
    <img
      src={src}
      alt=""
      width={16}
      height={16}
      className={cn("shrink-0 rounded", className)}
      loading="lazy"
      decoding="async"
      onError={() => setFailedSrc(src)}
    />
  );
}
