import { useEffect, useState } from "react";
import {
  promptPWAInstall,
  subscribePWAInstallAvailability,
} from "@/lib/pwa";

export function usePWAInstall() {
  const [isInstallAvailable, setIsInstallAvailable] = useState(false);

  useEffect(() => {
    return subscribePWAInstallAvailability(setIsInstallAvailable);
  }, []);

  return {
    isInstallAvailable,
    promptInstall: promptPWAInstall,
  };
}
