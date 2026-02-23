import { Menu } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useIsMobile } from "@/hooks/use-mobile";
import { useI18n } from "@/lib/i18n";
import { useUIStore } from "@/store/ui";

export function SidebarTrigger() {
  const { t } = useI18n();
  const isMobile = useIsMobile();
  const setSidebarOpen = useUIStore((s) => s.setSidebarOpen);

  if (!isMobile) return null;

  return (
    <Button
      variant="ghost"
      size="icon"
      className="shrink-0"
      onClick={() => setSidebarOpen(true)}
      aria-label={t("common.navigation")}
    >
      <Menu className="h-5 w-5" />
    </Button>
  );
}
