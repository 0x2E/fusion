import { createFileRoute, redirect } from "@tanstack/react-router";
import {
  defaultArticleFilter,
  isArticleFilter,
} from "@/lib/article-filter";
import { parsePositiveIntegerParam } from "@/lib/route-params";

export const Route = createFileRoute("/groups/$groupId/$filter")({
  beforeLoad: ({ params }) => {
    const groupId = parsePositiveIntegerParam(params.groupId);
    if (groupId === null) {
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
      to: "/groups/$groupId/$filter",
      params: {
        groupId: String(groupId),
        filter: defaultArticleFilter,
      },
      replace: true,
    });
  },
});
