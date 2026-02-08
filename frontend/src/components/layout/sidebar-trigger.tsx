import { Menu } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useIsMobile } from "@/hooks/use-mobile";
import { useUIStore } from "@/store/ui";

export function SidebarTrigger() {
  const isMobile = useIsMobile();
  const setSidebarOpen = useUIStore((s) => s.setSidebarOpen);

  if (!isMobile) return null;

  return (
    <Button
      variant="ghost"
      size="icon"
      className="shrink-0"
      onClick={() => setSidebarOpen(true)}
    >
      <Menu className="h-5 w-5" />
    </Button>
  );
}
