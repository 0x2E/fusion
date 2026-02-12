import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useI18n } from "@/lib/i18n";
import { useUIStore } from "@/store";

interface ShortcutItem {
  keys: string;
  action: string;
}

interface ShortcutSection {
  title: string;
  items: ShortcutItem[];
}

export function ShortcutsDialog() {
  const { t } = useI18n();
  const isShortcutsOpen = useUIStore((s) => s.isShortcutsOpen);
  const setShortcutsOpen = useUIStore((s) => s.setShortcutsOpen);

  const sections: ShortcutSection[] = [
    {
      title: t("shortcuts.section.global"),
      items: [
        { keys: "Cmd+K / Ctrl+K", action: t("shortcuts.action.toggleSearch") },
        { keys: "/", action: t("shortcuts.action.openSearch") },
        { keys: "Cmd+, / Ctrl+,", action: t("shortcuts.action.openSettings") },
        { keys: "?", action: t("shortcuts.action.showHelp") },
        { keys: "Esc", action: t("shortcuts.action.closeDialog") },
      ],
    },
    {
      title: t("shortcuts.section.article"),
      items: [
        {
          keys: "j / ArrowDown / n",
          action: t("shortcuts.action.nextArticle"),
        },
        {
          keys: "k / ArrowUp / p",
          action: t("shortcuts.action.previousArticle"),
        },
        { keys: "m", action: t("shortcuts.action.toggleRead") },
        { keys: "s / f", action: t("shortcuts.action.toggleStar") },
        { keys: "o / v", action: t("shortcuts.action.openOriginal") },
      ],
    },
    {
      title: t("shortcuts.section.navigation"),
      items: [
        { keys: "g u", action: t("shortcuts.action.goUnread") },
        { keys: "g a", action: t("shortcuts.action.goAll") },
        { keys: "g s", action: t("shortcuts.action.goStarred") },
        { keys: "g f", action: t("shortcuts.action.goFeeds") },
      ],
    },
  ];

  return (
    <Dialog open={isShortcutsOpen} onOpenChange={setShortcutsOpen}>
      <DialogContent className="sm:max-w-[560px] max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t("shortcuts.title")}</DialogTitle>
        </DialogHeader>
        <div className="space-y-6">
          {sections.map((section) => (
            <section key={section.title} className="space-y-3">
              <h3 className="text-sm font-semibold text-foreground">{section.title}</h3>
              <div className="space-y-2">
                {section.items.map((item) => (
                  <div
                    key={`${section.title}-${item.keys}`}
                    className="flex items-center justify-between gap-4 rounded-md border px-3 py-2"
                  >
                    <span className="text-sm text-muted-foreground">{item.action}</span>
                    <kbd className="rounded bg-muted px-2 py-1 font-mono text-xs font-medium text-foreground">
                      {item.keys}
                    </kbd>
                  </div>
                ))}
              </div>
            </section>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
