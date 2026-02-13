import DOMPurify from "dompurify";
import { resolveSafeExternalUrl, toSafeExternalUrl } from "@/lib/safe-url";

const purify = DOMPurify(window);

const ALLOWED_TAGS = [
  "p",
  "br",
  "strong",
  "em",
  "u",
  "s",
  "a",
  "img",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "ul",
  "ol",
  "li",
  "blockquote",
  "code",
  "pre",
  "table",
  "thead",
  "tbody",
  "tr",
  "th",
  "td",
  "div",
  "span",
  "hr",
  "figure",
  "figcaption",
];

const ALLOWED_ATTR = ["href", "src", "alt", "title", "class", "target", "rel"];

// Tags that are meaningful even when empty
const SELF_CLOSING_TAGS = new Set(["br", "hr", "img", "td", "th", "li"]);

const TRACKER_PATTERNS = [
  /feedburner/i,
  /doubleclick/i,
  /\/pixel[./]/i,
  /\/beacon[./]/i,
  /\/track[./]/i,
  /\/open[./]/i,
  /mail\.google\.com\/.*\/pixel/i,
  /emailtracking/i,
];

function isTrackingPixel(img: HTMLImageElement): boolean {
  const width = img.getAttribute("width");
  const height = img.getAttribute("height");
  if ((width === "1" || width === "0") && (height === "1" || height === "0")) {
    return true;
  }

  const src = img.getAttribute("src") || "";
  return TRACKER_PATTERNS.some((pattern) => pattern.test(src));
}

function isEmptyElement(el: Element): boolean {
  if (SELF_CLOSING_TAGS.has(el.tagName.toLowerCase())) return false;
  if (el.querySelector("img")) return false;
  const text = el.textContent || "";
  return text.trim().length === 0;
}

function sanitizeAnchors(root: DocumentFragment, articleUrl: string | null): void {
  for (const node of root.querySelectorAll("a")) {
    const href = node.getAttribute("href");
    const safeHref = resolveSafeExternalUrl(href, articleUrl);
    if (!safeHref) {
      node.removeAttribute("href");
      node.removeAttribute("target");
      node.removeAttribute("rel");
      continue;
    }

    node.setAttribute("href", safeHref);
    node.setAttribute("target", "_blank");
    node.setAttribute("rel", "noopener noreferrer");
  }
}

function sanitizeImages(root: DocumentFragment, articleUrl: string | null): void {
  for (const node of root.querySelectorAll("img")) {
    const img = node as HTMLImageElement;
    const safeSrc = resolveSafeExternalUrl(node.getAttribute("src"), articleUrl);
    if (!safeSrc) {
      img.remove();
      continue;
    }

    img.setAttribute("src", safeSrc);
    if (isTrackingPixel(img)) {
      img.remove();
    }
  }
}

function removeEmptyWrappers(root: DocumentFragment): void {
  const elements = Array.from(root.querySelectorAll("*")).reverse();
  for (const element of elements) {
    if (isEmptyElement(element)) {
      element.remove();
    }
  }
}

export function processArticleContent(
  html: string,
  articleUrl?: string,
): string {
  const safeArticleUrl = toSafeExternalUrl(articleUrl);
  const fragment = purify.sanitize(html, {
    ALLOWED_TAGS,
    ALLOWED_ATTR,
    ALLOW_DATA_ATTR: false,
    RETURN_DOM_FRAGMENT: true,
  }) as DocumentFragment;

  sanitizeAnchors(fragment, safeArticleUrl);
  sanitizeImages(fragment, safeArticleUrl);
  removeEmptyWrappers(fragment);

  const container = document.createElement("div");
  container.appendChild(fragment);
  return container.innerHTML;
}
