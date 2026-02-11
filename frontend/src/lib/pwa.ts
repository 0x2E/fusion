import { toast } from "sonner";

const SW_URL = "/sw.js";
const SKIP_WAITING_MESSAGE = { type: "SKIP_WAITING" } as const;

let deferredInstallPrompt: BeforeInstallPromptEvent | null = null;
let shouldReloadOnControllerChange = false;
const installAvailabilityListeners = new Set<(isAvailable: boolean) => void>();

function notifyInstallAvailability(): void {
  const isAvailable = deferredInstallPrompt !== null;
  for (const listener of installAvailabilityListeners) {
    listener(isAvailable);
  }
}

function showUpdateToast(registration: ServiceWorkerRegistration): void {
  toast.info("Update available", {
    id: "pwa-update",
    description: "Reload to use the latest Fusion version.",
    duration: Number.POSITIVE_INFINITY,
    action: {
      label: "Reload",
      onClick: () => {
        shouldReloadOnControllerChange = true;
        if (registration.waiting) {
          registration.waiting.postMessage(SKIP_WAITING_MESSAGE);
          return;
        }

        window.location.reload();
      },
    },
  });
}

function setupServiceWorkerUpdateFlow(registration: ServiceWorkerRegistration): void {
  if (registration.waiting) {
    showUpdateToast(registration);
  }

  registration.addEventListener("updatefound", () => {
    const installing = registration.installing;
    if (!installing) {
      return;
    }

    installing.addEventListener("statechange", () => {
      if (installing.state === "installed" && navigator.serviceWorker.controller) {
        showUpdateToast(registration);
      }
    });
  });
}

export function isPWAInstallAvailable(): boolean {
  return deferredInstallPrompt !== null;
}

export function subscribePWAInstallAvailability(
  listener: (isAvailable: boolean) => void,
): () => void {
  installAvailabilityListeners.add(listener);
  listener(isPWAInstallAvailable());

  return () => {
    installAvailabilityListeners.delete(listener);
  };
}

export async function promptPWAInstall(): Promise<boolean> {
  if (!deferredInstallPrompt) {
    return false;
  }

  const promptEvent = deferredInstallPrompt;
  deferredInstallPrompt = null;
  notifyInstallAvailability();

  await promptEvent.prompt();
  const { outcome } = await promptEvent.userChoice;
  return outcome === "accepted";
}

export function registerPWA(): void {
  if (import.meta.env.DEV || !("serviceWorker" in navigator)) {
    return;
  }

  navigator.serviceWorker.addEventListener("controllerchange", () => {
    if (shouldReloadOnControllerChange) {
      window.location.reload();
    }
  });

  window.addEventListener("beforeinstallprompt", (event) => {
    event.preventDefault();
    deferredInstallPrompt = event;
    notifyInstallAvailability();
  });

  window.addEventListener("appinstalled", () => {
    deferredInstallPrompt = null;
    notifyInstallAvailability();
    toast.success("Fusion installed");
  });

  window.addEventListener("load", () => {
    void navigator.serviceWorker
      .register(SW_URL)
      .then((registration) => {
        setupServiceWorkerUpdateFlow(registration);
      })
      .catch((error: unknown) => {
        console.error("Failed to register service worker", error);
      });
  });
}
