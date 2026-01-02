import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Separator } from "@/components/ui/separator";
import { useAuthStore } from "@/store/auth";
import { LogOut, Monitor, Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";

interface SettingsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function SettingsDialog({ open, onOpenChange }: SettingsDialogProps) {
  const { theme, setTheme } = useTheme();
  const { logout } = useAuthStore();

  const handleLogout = async () => {
    try {
      await logout();
      onOpenChange(false);
    } catch (error) {
      // Error already handled in store
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Settings</DialogTitle>
        </DialogHeader>

        <div className="py-4 space-y-6">
          {/* Theme Setting */}
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-medium mb-3">Appearance</h3>
              <RadioGroup value={theme} onValueChange={setTheme}>
                <div className="flex items-center space-x-2 py-2">
                  <RadioGroupItem value="light" id="theme-light" />
                  <Label
                    htmlFor="theme-light"
                    className="flex items-center gap-2 cursor-pointer"
                  >
                    <Sun className="w-4 h-4" />
                    <span>Light</span>
                  </Label>
                </div>
                <div className="flex items-center space-x-2 py-2">
                  <RadioGroupItem value="dark" id="theme-dark" />
                  <Label
                    htmlFor="theme-dark"
                    className="flex items-center gap-2 cursor-pointer"
                  >
                    <Moon className="w-4 h-4" />
                    <span>Dark</span>
                  </Label>
                </div>
                <div className="flex items-center space-x-2 py-2">
                  <RadioGroupItem value="system" id="theme-system" />
                  <Label
                    htmlFor="theme-system"
                    className="flex items-center gap-2 cursor-pointer"
                  >
                    <Monitor className="w-4 h-4" />
                    <span>System</span>
                  </Label>
                </div>
              </RadioGroup>
            </div>
          </div>

          <Separator />

          {/* Account Section */}
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-medium mb-3">Account</h3>
              <Button
                variant="destructive"
                className="w-full justify-start"
                onClick={handleLogout}
              >
                <LogOut className="w-4 h-4 mr-2" />
                Logout
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
