import type { ComponentProps } from "react";
import { cn } from "@/lib/utils";

type ContentHeaderProps = ComponentProps<"header">;

export function ContentHeader({ className, ...props }: ContentHeaderProps) {
  return (
    <header
      className={cn(
        "flex h-14 shrink-0 items-center justify-between border-b px-4 sm:px-6",
        className,
      )}
      {...props}
    />
  );
}
