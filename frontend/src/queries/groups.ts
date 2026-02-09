import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { groupAPI, type Group } from "@/lib/api";
import { queryKeys } from "./keys";

export const groupQueries = {
  list: () =>
    queryOptions({
      queryKey: queryKeys.groups.list(),
      queryFn: async () => {
        const res = await groupAPI.list();
        return res.data;
      },
    }),
};

export function useGroups() {
  return useQuery(groupQueries.list());
}

export function useCreateGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (name: string) => {
      const res = await groupAPI.create({ name });
      return res.data!;
    },
    onSuccess: (group) => {
      qc.setQueryData(queryKeys.groups.list(), (old: Group[] | undefined) =>
        old ? [...old, group] : [group],
      );
    },
  });
}

export function useUpdateGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, name }: { id: number; name: string }) => {
      await groupAPI.update(id, { name });
      return { id, name };
    },
    onSuccess: ({ id, name }) => {
      qc.setQueryData(queryKeys.groups.list(), (old: Group[] | undefined) =>
        old?.map((g) => (g.id === id ? { ...g, name } : g)),
      );
    },
  });
}

export function useDeleteGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      await groupAPI.delete(id);
      return id;
    },
    onSuccess: (id) => {
      qc.setQueryData(queryKeys.groups.list(), (old: Group[] | undefined) =>
        old?.filter((g) => g.id !== id),
      );
      qc.invalidateQueries({ queryKey: queryKeys.feeds.all });
    },
  });
}
