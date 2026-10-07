import { beforeEach, describe, expect, it } from "vitest";
import {
  articleListKey,
  useArticlePinsStore,
} from "./article-pins";

// The store is module-level state; each test starts from clean pins.
beforeEach(() => {
  useArticlePinsStore.setState({ pins: {} });
});

describe("articleListKey", () => {
  it("distinguishes feed, group, and filter dimensions", () => {
    expect(articleListKey(1, null, "unread")).not.toBe(
      articleListKey(2, null, "unread"),
    );
    expect(articleListKey(null, 1, "unread")).not.toBe(
      articleListKey(null, 2, "unread"),
    );
    expect(articleListKey(1, null, "unread")).not.toBe(
      articleListKey(1, null, "all"),
    );
  });

  it("treats null and 0 as the same dimension value", () => {
    expect(articleListKey(null, null, "all")).toBe(
      articleListKey(0, 0, "all"),
    );
  });
});

describe("pinItems", () => {
  it("pins ids under the given list key only", () => {
    const store = useArticlePinsStore.getState();
    store.pinItems("a", [1, 2]);

    expect(useArticlePinsStore.getState().pins["a"]).toEqual(
      new Set([1, 2]),
    );
    expect(useArticlePinsStore.getState().pins["b"]).toBeUndefined();
  });

  it("ignores non-positive ids", () => {
    useArticlePinsStore.getState().pinItems("a", [0, -3, 5]);

    expect(useArticlePinsStore.getState().pins["a"]).toEqual(new Set([5]));
  });

  it("is a no-op for an empty id list", () => {
    useArticlePinsStore.getState().pinItems("a", [1]);
    const before = useArticlePinsStore.getState().pins;

    useArticlePinsStore.getState().pinItems("a", []);

    // Same reference: no state change was emitted.
    expect(useArticlePinsStore.getState().pins).toBe(before);
  });

  it("is a no-op when every id is already pinned", () => {
    useArticlePinsStore.getState().pinItems("a", [1, 2]);
    const before = useArticlePinsStore.getState().pins;

    useArticlePinsStore.getState().pinItems("a", [1, 2, 2]);

    expect(useArticlePinsStore.getState().pins).toBe(before);
  });
});

describe("unpinItems", () => {
  it("removes pinned ids", () => {
    useArticlePinsStore.getState().pinItems("a", [1, 2, 3]);
    useArticlePinsStore.getState().unpinItems("a", [2, 3]);

    expect(useArticlePinsStore.getState().pins["a"]).toEqual(new Set([1]));
  });

  it("is a no-op for an unknown key", () => {
    expect(() =>
      useArticlePinsStore.getState().unpinItems("missing", [1]),
    ).not.toThrow();
    expect(useArticlePinsStore.getState().pins).toEqual({});
  });
});

describe("clearPins", () => {
  it("drops the whole key, leaving other lists intact", () => {
    useArticlePinsStore.getState().pinItems("a", [1]);
    useArticlePinsStore.getState().pinItems("b", [2]);
    useArticlePinsStore.getState().clearPins("a");

    expect(useArticlePinsStore.getState().pins).toEqual({
      b: new Set([2]),
    });
  });

  it("is a no-op for an unknown key", () => {
    const before = useArticlePinsStore.getState().pins;
    useArticlePinsStore.getState().clearPins("missing");
    expect(useArticlePinsStore.getState().pins).toBe(before);
  });
});
