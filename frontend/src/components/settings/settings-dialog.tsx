import { useState } from "react";
import { useTheme } from "next-themes";
import { toast } from "sonner";
import { Bug, Download, Github, Info, Palette } from "lucide-react";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  articlePageSizeOptions,
  supportedLocales,
  usePreferencesStore,
  useUIStore,
} from "@/store";
import { usePWAInstall } from "@/hooks/use-pwa-install";
import { localeLabels, useI18n } from "@/lib/i18n";
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
        "flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors sm:w-full sm:gap-2.5 sm:px-2",
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
  const { t } = useI18n();
  const { theme, setTheme } = useTheme();
  const { locale, articlePageSize, setLocale, setArticlePageSize } =
    usePreferencesStore();

  return (
    <div className="space-y-5">
      {/* Language */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <p className="text-sm font-medium">{t("settings.language.label")}</p>
          <p className="text-[13px] text-muted-foreground">
            {t("settings.language.description")}
          </p>
        </div>
        <Select value={locale} onValueChange={setLocale}>
          <SelectTrigger className="w-auto gap-2 border-border">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {supportedLocales.map((localeCode) => (
              <SelectItem key={localeCode} value={localeCode}>
                {localeLabels[localeCode] ?? localeCode}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Articles per load */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <p className="text-sm font-medium">
            {t("settings.articlePageSize.label")}
          </p>
          <p className="text-[13px] text-muted-foreground">
            {t("settings.articlePageSize.description")}
          </p>
        </div>
        <Select
          value={articlePageSize.toString()}
          onValueChange={(value) => {
            const parsed = Number.parseInt(value, 10);
            if (!Number.isNaN(parsed)) {
              setArticlePageSize(parsed);
            }
          }}
        >
          <SelectTrigger className="w-auto gap-2 border-border">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {articlePageSizeOptions.map((size) => (
              <SelectItem key={size} value={size.toString()}>
                {size}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Theme */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <p className="text-sm font-medium">{t("settings.theme.label")}</p>
          <p className="text-[13px] text-muted-foreground">
            {t("settings.theme.description")}
          </p>
        </div>
        <Select value={theme} onValueChange={setTheme}>
          <SelectTrigger className="w-auto gap-2 border-border">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="light">{t("settings.theme.light")}</SelectItem>
            <SelectItem value="dark">{t("settings.theme.dark")}</SelectItem>
            <SelectItem value="system">{t("settings.theme.system")}</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  );
}

function AboutContent() {
  const { t } = useI18n();
  const { isInstallAvailable, promptInstall } = usePWAInstall();
  const [isInstalling, setIsInstalling] = useState(false);

  const handleInstall = async () => {
    setIsInstalling(true);
    try {
      const isInstalled = await promptInstall();
      if (!isInstalled) {
        toast.info(t("settings.installCancelled"));
      }
    } finally {
      setIsInstalling(false);
    }
  };

  return (
    <div className="flex h-full flex-col items-center justify-center gap-4 pb-8">
      <img
        src="/icon-96.png"
        alt={t("common.fusionLogo")}
        className="h-16 w-16 rounded-2xl"
      />
      <div className="text-center">
        <h3 className="text-xl font-semibold">Fusion</h3>
        <p className="mt-1 text-xs text-muted-foreground">{__APP_VERSION__}</p>
        <p className="mt-1.5 text-sm text-muted-foreground">
          {t("settings.about.description")}
        </p>
      </div>
      <div className="flex gap-2">
        {isInstallAvailable && (
          <Button
            variant="default"
            size="sm"
            onClick={() => {
              void handleInstall();
            }}
            disabled={isInstalling}
          >
            <Download className="h-4 w-4" />
            {isInstalling
              ? t("settings.about.installing")
              : t("settings.about.install")}
          </Button>
        )}
        <Button variant="outline" size="sm" asChild>
          <a
            href="https://github.com/0x2e/fusion"
            target="_blank"
            rel="noopener noreferrer"
          >
            <Github className="h-4 w-4" />
            {t("settings.about.github")}
          </a>
        </Button>
        <Button variant="outline" size="sm" asChild>
          <a
            href="https://github.com/0x2e/fusion/issues"
            target="_blank"
            rel="noopener noreferrer"
          >
            <Bug className="h-4 w-4" />
            {t("settings.about.reportIssue")}
          </a>
        </Button>
      </div>
      <p className="mt-auto text-xs text-muted-foreground">
        {t("settings.about.license")}
      </p>
    </div>
  );
}

export function SettingsDialog() {
  const { t } = useI18n();
  const { isSettingsOpen, setSettingsOpen } = useUIStore();
  const [activeTab, setActiveTab] = useState<SettingsTab>("appearance");

  const tabTitles: Record<SettingsTab, string> = {
    appearance: t("settings.tab.appearance"),
    about: t("settings.tab.about"),
  };

  return (
    <Dialog open={isSettingsOpen} onOpenChange={setSettingsOpen}>
      <DialogContent className="flex max-h-[85vh] flex-col sm:flex-row h-auto sm:h-[560px] sm:max-w-4xl gap-0 overflow-hidden p-0">
        {/* Sidebar (desktop) / Tab bar (mobile) */}
        <div className="flex shrink-0 flex-row border-b border-border bg-muted/30 px-3 pt-3 sm:w-[200px] sm:flex-col sm:border-b-0 sm:border-r sm:pt-4">
          <h2 className="hidden px-2 text-sm font-semibold sm:block">
            {t("common.settings")}
          </h2>
          <nav className="flex gap-0.5 sm:mt-2 sm:flex-col">
            <NavItem
              icon={<Palette className="h-4 w-4" />}
              label={t("settings.tab.appearance")}
              active={activeTab === "appearance"}
              onClick={() => setActiveTab("appearance")}
            />
            <NavItem
              icon={<Info className="h-4 w-4" />}
              label={t("settings.tab.about")}
              active={activeTab === "about"}
              onClick={() => setActiveTab("about")}
            />
          </nav>
        </div>

        {/* Content */}
        <div className="flex min-h-0 flex-1 flex-col overflow-hidden p-5 sm:p-6">
          <h2 className="mb-4 shrink-0 text-lg font-semibold sm:mb-6">
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
