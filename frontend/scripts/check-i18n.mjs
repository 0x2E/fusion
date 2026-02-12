import fs from "node:fs";
import path from "node:path";
import process from "node:process";

const MESSAGES_DIR = path.resolve("src/lib/i18n/messages");
const EN_FILE = path.join(MESSAGES_DIR, "en.ts");
const LOCALES = ["zh", "de", "fr", "es", "ru", "pt", "sv"];

function extractKeys(fileContent) {
  const keyPattern = /"([^"\\]+)"\s*:/g;
  const keys = [];

  for (const match of fileContent.matchAll(keyPattern)) {
    keys.push(match[1]);
  }

  return keys;
}

function printMismatch(type, locale, keys) {
  console.error(`[${locale}] ${type} (${keys.length})`);
  for (const key of keys) {
    console.error(`  - ${key}`);
  }
}

const enSource = fs.readFileSync(EN_FILE, "utf8");
const enKeys = extractKeys(enSource);
const enSet = new Set(enKeys);

let hasError = false;

for (const locale of LOCALES) {
  const localeFile = path.join(MESSAGES_DIR, `${locale}.ts`);
  if (!fs.existsSync(localeFile)) {
    hasError = true;
    console.error(`[${locale}] Missing locale file: ${localeFile}`);
    continue;
  }

  const localeSource = fs.readFileSync(localeFile, "utf8");
  const localeKeys = extractKeys(localeSource);
  const localeSet = new Set(localeKeys);

  const missing = enKeys.filter((key) => !localeSet.has(key));
  const extra = localeKeys.filter((key) => !enSet.has(key));

  if (missing.length > 0) {
    hasError = true;
    printMismatch("Missing keys", locale, missing);
  }

  if (extra.length > 0) {
    hasError = true;
    printMismatch("Unknown keys", locale, extra);
  }

  if (missing.length === 0 && extra.length === 0) {
    console.log(`[${locale}] OK (${localeKeys.length} keys)`);
  }
}

if (hasError) {
  process.exitCode = 1;
} else {
  console.log("All locale dictionaries are complete.");
}
