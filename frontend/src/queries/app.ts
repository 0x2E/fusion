import { queryOptions, useQuery } from "@tanstack/react-query";
import { appAPI } from "@/lib/api";
import { queryKeys } from "./keys";

export const appQueries = {
  info: () =>
    queryOptions({
      queryKey: queryKeys.app.info(),
      queryFn: async () => {
        const res = await appAPI.getInfo();
        return res.data!;
      },
    }),
};

export function useAppInfo() {
  return useQuery(appQueries.info());
}
