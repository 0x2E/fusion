import { describe, expect, it } from "vitest";
import { getFeedErrorPreview } from "./feed-error";

describe("getFeedErrorPreview", () => {
  it("collapses whitespace before measuring length", () => {
    expect(
      getFeedErrorPreview("line one\nline two\t\twith   spaces"),
    ).toBe("line one line two with spaces");
  });

  it("passes short messages through unchanged", () => {
    expect(getFeedErrorPreview("404 Not Found")).toBe("404 Not Found");
  });

  it("truncates at exactly 48 characters with an ellipsis", () => {
    const exactly48 = "a".repeat(48);
    expect(getFeedErrorPreview(exactly48)).toBe(exactly48);

    const long = "x".repeat(60);
    expect(getFeedErrorPreview(long)).toBe(`${"x".repeat(48)}...`);
  });

  it("measures the normalized form, not the raw input", () => {
    // 100 raw characters that normalize to 10 must not be truncated.
    expect(getFeedErrorPreview("\n".repeat(100) + "short")).toBe("short");
  });
});
