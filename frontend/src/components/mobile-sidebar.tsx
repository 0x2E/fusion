import { Sidebar } from "@/components/sidebar";
import { Sheet, SheetContent } from "@/components/ui/sheet";

interface MobileSidebarProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSearchOpen: () => void;
  onSettingsOpen: () => void;
}

export function MobileSidebar({
  open,
  onOpenChange,
  onSearchOpen,
  onSettingsOpen,
}: MobileSidebarProps) {
  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="left" className="p-0 w-64 h-full">
        <Sidebar onSearchOpen={onSearchOpen} onSettingsOpen={onSettingsOpen} />
      </SheetContent>
    </Sheet>
  );
}
