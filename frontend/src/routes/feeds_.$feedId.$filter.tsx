import { createFileRoute, redirect } from "@tanstack/react-router";
import {
  defaultArticleFilter,
  isArticleFilter,
} from "@/lib/article-filter";
import { parsePositiveIntegerParam } from "@/lib/route-params";

export const Route = createFileRoute("/feeds_/$feedId/$filter")({
  beforeLoad: ({ params }) => {
    const feedId = parsePositiveIntegerParam(params.feedId);
    if (feedId === null) {
      throw redirect({
        to: "/$filter",
        params: { filter: defaultArticleFilter },
        replace: true,
      });
    }

    if (isArticleFilter(params.filter)) {
      return;
    }

    throw redirect({
      to: "/feeds/$feedId/$filter",
      params: {
        feedId: String(feedId),
        filter: defaultArticleFilter,
      },
      replace: true,
    });
  },
});
