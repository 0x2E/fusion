import DOMPurify from "dompurify";

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

function isAbsoluteUrl(url: string): boolean {
  return /^(https?:\/\/|\/\/|data:|mailto:)/i.test(url);
}

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

function getBaseUrl(url: string): string {
  try {
    const parsed = new URL(url);
    const pathParts = parsed.pathname.split("/");
    pathParts.pop(); // Remove filename
    return parsed.origin + pathParts.join("/") + "/";
  } catch {
    return "";
  }
}

function resolveUrl(url: string, articleUrl: string): string {
  if (isAbsoluteUrl(url)) return url;
  try {
    return new URL(url, getBaseUrl(articleUrl)).href;
  } catch {
    return url;
  }
}

// Process links and images after attribute sanitization
purify.addHook("afterSanitizeAttributes", (node) => {
  if (node.tagName === "A") {
    node.setAttribute("target", "_blank");
    node.setAttribute("rel", "noopener noreferrer");

    const href = node.getAttribute("href");
    if (href && currentArticleUrl && !isAbsoluteUrl(href)) {
      node.setAttribute("href", resolveUrl(href, currentArticleUrl));
    }
  }

  if (node.tagName === "IMG") {
    const img = node as HTMLImageElement;

    // Resolve relative src
    const src = img.getAttribute("src");
    if (src && currentArticleUrl && !isAbsoluteUrl(src)) {
      img.setAttribute("src", resolveUrl(src, currentArticleUrl));
    }

    // Remove tracking pixels
    if (isTrackingPixel(img)) {
      img.remove();
    }
  }
});

// Remove empty wrapper elements after sanitization
purify.addHook("afterSanitizeElements", (node) => {
  if (node.nodeType === Node.ELEMENT_NODE && isEmptyElement(node as Element)) {
    node.parentNode?.removeChild(node);
  }
});

// Module-level state to pass article URL into hooks
let currentArticleUrl: string | undefined;

export function processArticleContent(
  html: string,
  articleUrl?: string,
): string {
  currentArticleUrl = articleUrl;
  const result = purify.sanitize(html, {
    ALLOWED_TAGS,
    ALLOWED_ATTR,
    ALLOW_DATA_ATTR: false,
  });
  currentArticleUrl = undefined;
  return result;
}
