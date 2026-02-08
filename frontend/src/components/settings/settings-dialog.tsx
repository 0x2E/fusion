import { useState } from "react";
import { useTheme } from "next-themes";
import { Github, Info, Palette } from "lucide-react";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useUIStore } from "@/store";
import { cn } from "@/lib/utils";

type SettingsTab = "appearance" | "about";

interface NavItemProps {
  icon: React.ReactNode;
  label: string;
  active: boolean;
  onClick: () => void;
}

function NavItem({ icon, label, active, onClick }: NavItemProps) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "flex w-full items-center gap-2.5 rounded-md px-2 py-2 text-sm transition-colors",
        active
          ? "bg-accent font-medium text-foreground"
          : "text-muted-foreground hover:bg-accent/50 hover:text-foreground",
      )}
    >
      {icon}
      <span className="whitespace-nowrap">{label}</span>
    </button>
  );
}

function AppearanceContent() {
  const { theme, setTheme } = useTheme();

  return (
    <div className="space-y-5">
      {/* Language */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <p className="text-sm font-medium">Language</p>
          <p className="text-[13px] text-muted-foreground">
            Select your preferred language
          </p>
        </div>
        <Select defaultValue="en">
          <SelectTrigger className="w-auto gap-2 border-border">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="en">English</SelectItem>
            <SelectItem value="zh">中文</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Theme */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <p className="text-sm font-medium">Theme</p>
          <p className="text-[13px] text-muted-foreground">
            Choose your color theme
          </p>
        </div>
        <Select value={theme} onValueChange={setTheme}>
          <SelectTrigger className="w-auto gap-2 border-border">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="light">Light</SelectItem>
            <SelectItem value="dark">Dark</SelectItem>
            <SelectItem value="system">System</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  );
}

function AboutContent() {
  return (
    <div className="flex flex-col items-center gap-4">
      <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary">
        <span className="text-2xl font-bold text-primary-foreground">F</span>
      </div>
      <div className="text-center">
        <h3 className="text-xl font-semibold">Fusion</h3>
        <p className="mt-1 text-sm text-muted-foreground">Version 0.1.0</p>
      </div>
      <Button variant="outline" asChild className="mt-2">
        <a href="https://github.com/0x2e/fusion" target="_blank" rel="noopener noreferrer">
          <Github className="h-4 w-4" />
          GitHub
        </a>
      </Button>
    </div>
  );
}

const tabTitles: Record<SettingsTab, string> = {
  appearance: "Appearance",
  about: "About",
};

export function SettingsDialog() {
  const { isSettingsOpen, setSettingsOpen } = useUIStore();
  const [activeTab, setActiveTab] = useState<SettingsTab>("appearance");

  return (
    <Dialog open={isSettingsOpen} onOpenChange={setSettingsOpen}>
      <DialogContent className="flex max-h-[85vh] h-auto sm:h-[560px] sm:max-w-4xl gap-0 overflow-hidden p-0">
        {/* Sidebar */}
        <div className="flex w-[200px] shrink-0 flex-col border-r border-border bg-muted/30 p-3 pt-4">
          <h2 className="px-2 text-sm font-semibold">Settings</h2>
          <nav className="mt-2 space-y-0.5">
            <NavItem
              icon={<Palette className="h-4 w-4" />}
              label="Appearance"
              active={activeTab === "appearance"}
              onClick={() => setActiveTab("appearance")}
            />
            <NavItem
              icon={<Info className="h-4 w-4" />}
              label="About"
              active={activeTab === "about"}
              onClick={() => setActiveTab("about")}
            />
          </nav>
        </div>

        {/* Content */}
        <div className="flex flex-1 flex-col overflow-hidden p-6">
          <h2 className="mb-6 shrink-0 text-lg font-semibold">
            {tabTitles[activeTab]}
          </h2>

          <div className="flex-1 overflow-y-auto">
            {activeTab === "appearance" && <AppearanceContent />}
            {activeTab === "about" && <AboutContent />}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
