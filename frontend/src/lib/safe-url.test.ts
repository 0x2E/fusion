import { describe, expect, it } from "vitest";
import {
  resolveSafeExternalUrl,
  toSafeExternalUrl,
} from "./safe-url";

describe("toSafeExternalUrl", () => {
  it("returns the normalized href for http and https URLs", () => {
    expect(toSafeExternalUrl("https://example.com/a?b=1")).toBe(
      "https://example.com/a?b=1",
    );
    expect(toSafeExternalUrl("  HTTP://Example.COM/Path  ")).toBe(
      "http://example.com/Path",
    );
  });

  it("rejects non-http(s) protocols", () => {
    expect(toSafeExternalUrl("javascript:alert(1)")).toBeNull();
    expect(toSafeExternalUrl("data:text/html,<script>")).toBeNull();
    expect(toSafeExternalUrl("ftp://example.com/file")).toBeNull();
    expect(toSafeExternalUrl("mailto:user@example.com")).toBeNull();
  });

  it("rejects unparseable and empty input", () => {
    expect(toSafeExternalUrl("")).toBeNull();
    expect(toSafeExternalUrl("   ")).toBeNull();
    expect(toSafeExternalUrl("not a url")).toBeNull();
    // Protocol-relative URLs have no base to resolve against here.
    expect(toSafeExternalUrl("//example.com")).toBeNull();
    expect(toSafeExternalUrl(null)).toBeNull();
    expect(toSafeExternalUrl(undefined)).toBeNull();
  });
});

describe("resolveSafeExternalUrl", () => {
  it("resolves a relative URL against a safe base", () => {
    expect(
      resolveSafeExternalUrl("/articles/1", "https://example.com/feed"),
    ).toBe("https://example.com/articles/1");
  });

  it("keeps an absolute raw URL even when the base is missing or unsafe", () => {
    // URL normalization adds the empty root path.
    expect(resolveSafeExternalUrl("https://other.com", undefined)).toBe(
      "https://other.com/",
    );
    // An unsafe base is discarded, not used.
    expect(resolveSafeExternalUrl("https://other.com", "javascript:x")).toBe(
      "https://other.com/",
    );
  });

  it("rejects unsafe results", () => {
    expect(
      resolveSafeExternalUrl("javascript:alert(1)", "https://example.com"),
    ).toBeNull();
    // Relative raw + unusable base cannot produce an absolute URL.
    expect(resolveSafeExternalUrl("/path", "javascript:x")).toBeNull();
    expect(resolveSafeExternalUrl(null, "https://example.com")).toBeNull();
  });
});
