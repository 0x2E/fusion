import { createFileRoute, redirect } from "@tanstack/react-router";
import { defaultArticleFilter } from "@/lib/article-filter";
import { parsePositiveIntegerParam } from "@/lib/route-params";

export const Route = createFileRoute("/groups/$groupId")({
  beforeLoad: ({ params, location }) => {
    const groupId = parsePositiveIntegerParam(params.groupId);
    if (groupId === null) {
      throw redirect({
        to: "/$filter",
        params: { filter: defaultArticleFilter },
        replace: true,
      });
    }

    const currentPath = location.pathname.replace(/\/+$/, "") || "/";
    if (currentPath !== `/groups/${params.groupId}`) {
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
