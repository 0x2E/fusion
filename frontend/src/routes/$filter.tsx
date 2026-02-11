import { createFileRoute, redirect } from "@tanstack/react-router";
import {
  defaultArticleFilter,
  isArticleFilter,
} from "@/lib/article-filter";

export const Route = createFileRoute("/$filter")({
  beforeLoad: ({ params }) => {
    if (isArticleFilter(params.filter)) {
      return;
    }

    throw redirect({
      to: "/$filter",
      params: { filter: defaultArticleFilter },
      replace: true,
    });
  },
});
