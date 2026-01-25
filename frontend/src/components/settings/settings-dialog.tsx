import { useTheme } from "next-themes";
import { Monitor, Moon, Sun } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { useUIStore } from "@/store";
import { cn } from "@/lib/utils";

export function SettingsDialog() {
  const { isSettingsOpen, setSettingsOpen } = useUIStore();
  const { theme, setTheme } = useTheme();

  return (
    <Dialog open={isSettingsOpen} onOpenChange={setSettingsOpen}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Settings</DialogTitle>
        </DialogHeader>

        <div className="space-y-6">
          {/* Appearance section */}
          <div className="space-y-4">
            <h3 className="text-sm font-medium">Appearance</h3>

            {/* Language */}
            <div className="flex items-center justify-between">
              <div>
                <label className="text-sm">Language</label>
                <p className="text-xs text-muted-foreground">
                  Select your preferred language
                </p>
              </div>
              <Select defaultValue="en">
                <SelectTrigger className="w-[140px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="en">English</SelectItem>
                  <SelectItem value="zh">中文</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Theme */}
            <div className="space-y-2">
              <div>
                <label className="text-sm">Theme</label>
                <p className="text-xs text-muted-foreground">
                  Select your preferred theme
                </p>
              </div>
              <div className="grid grid-cols-3 gap-2">
                <button
                  onClick={() => setTheme("light")}
                  className={cn(
                    "flex flex-col items-center gap-2 rounded-lg border p-3 transition-colors",
                    theme === "light"
                      ? "border-primary bg-accent"
                      : "border-border hover:bg-accent/50",
                  )}
                >
                  <Sun className="h-5 w-5" />
                  <span className="text-xs">Light</span>
                </button>
                <button
                  onClick={() => setTheme("dark")}
                  className={cn(
                    "flex flex-col items-center gap-2 rounded-lg border p-3 transition-colors",
                    theme === "dark"
                      ? "border-primary bg-accent"
                      : "border-border hover:bg-accent/50",
                  )}
                >
                  <Moon className="h-5 w-5" />
                  <span className="text-xs">Dark</span>
                </button>
                <button
                  onClick={() => setTheme("system")}
                  className={cn(
                    "flex flex-col items-center gap-2 rounded-lg border p-3 transition-colors",
                    theme === "system"
                      ? "border-primary bg-accent"
                      : "border-border hover:bg-accent/50",
                  )}
                >
                  <Monitor className="h-5 w-5" />
                  <span className="text-xs">System</span>
                </button>
              </div>
            </div>
          </div>

          <Separator />

          {/* About section */}
          <div className="space-y-2">
            <h3 className="text-sm font-medium">About</h3>
            <div className="rounded-lg border p-3">
              <div className="flex items-center gap-2">
                <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary">
                  <span className="text-sm font-bold text-primary-foreground">
                    F
                  </span>
                </div>
                <div>
                  <p className="text-sm font-medium">Fusion</p>
                  <p className="text-xs text-muted-foreground">
                    A modern RSS reader
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
