import { useQuery } from "@tanstack/react-query";
import { getRouteApi } from "@tanstack/react-router";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
  DISABLED_ROW_DESKTOP,
  DISABLED_ROW_MOBILE,
  DataTablePage,
  useDataTable,
} from "@/components/data-table";
import { useMediaQuery } from "@/hooks";
import { useTableUrlState } from "@/hooks/use-table-url-state";
import { getInvitations, searchInvitations } from "../api";
import { INVITATION_STATUS, getInvitationStatusOptions } from "../constants";
import { isInvitationExpired } from "../lib/utils";
import type { Invitation } from "../types";
import { DataTableBulkActions } from "./data-table-bulk-actions";
import { useInvitationsColumns } from "./invitations-columns";
import { InvitationsMobileList } from "./invitations-mobile-list";

const route = getRouteApi("/_authenticated/invitation-codes/");

function isDisabledInvitationRow(item: Invitation) {
  return (
    item.status !== INVITATION_STATUS.ENABLED ||
    isInvitationExpired(item.expired_time, item.status)
  );
}

export function InvitationsTable() {
  const { t } = useTranslation();
  const columns = useInvitationsColumns();
  const isMobile = useMediaQuery("(max-width: 640px)");

  const {
    globalFilter,
    onGlobalFilterChange,
    columnFilters,
    onColumnFiltersChange,
    pagination,
    onPaginationChange,
    ensurePageInRange,
  } = useTableUrlState({
    search: route.useSearch(),
    navigate: route.useNavigate(),
    pagination: { defaultPage: 1, defaultPageSize: isMobile ? 10 : 20 },
    globalFilter: { enabled: true, key: "filter" },
    columnFilters: [{ columnId: "status", searchKey: "status", type: "array" }],
  });

  const statusFilter =
    (columnFilters.find((filter) => filter.id === "status")?.value as
      string[] | undefined) ?? [];
  const statusFilterValue = statusFilter[0] ?? "";

  const { data, isLoading, isFetching } = useQuery({
    queryKey: [
      "invitations",
      pagination.pageIndex + 1,
      pagination.pageSize,
      globalFilter,
      statusFilterValue,
    ],
    queryFn: async () => {
      const hasFilter = globalFilter?.trim();
      const hasStatusFilter = statusFilterValue !== "";
      const params = {
        p: pagination.pageIndex + 1,
        page_size: pagination.pageSize,
      };

      const result =
        hasFilter || hasStatusFilter
          ? await searchInvitations({
              ...params,
              keyword: globalFilter,
              status: statusFilterValue,
            })
          : await getInvitations(params);

      if (!result.success) {
        toast.error(result.message || t("Failed to load invitation codes"));
        return { items: [], total: 0 };
      }

      return {
        items: result.data?.items || [],
        total: result.data?.total || 0,
      };
    },
    placeholderData: (previousData) => previousData,
  });

  const invitations = data?.items || [];

  const { table } = useDataTable({
    data: invitations,
    columns,
    enableRowSelection: true,
    columnFilters,
    globalFilter,
    pagination,
    globalFilterFn: (row, _columnId, filterValue) => {
      const name = String(row.getValue("name")).toLowerCase();
      const id = String(row.getValue("id"));
      const searchValue = String(filterValue).toLowerCase();
      return name.includes(searchValue) || id.includes(searchValue);
    },
    onPaginationChange,
    onGlobalFilterChange,
    onColumnFiltersChange,
    manualPagination: true,
    manualFiltering: true,
    totalCount: data?.total || 0,
    ensurePageInRange,
  });

  const invitationStatusOptions = useMemo(
    () => getInvitationStatusOptions(t),
    [t]
  );

  return (
    <DataTablePage
      table={table}
      columns={columns}
      isLoading={isLoading}
      isFetching={isFetching}
      emptyTitle={t("No invitation codes found")}
      emptyDescription={t(
        "No invitation codes available. Generate your first invitation code to get started."
      )}
      skeletonKeyPrefix="invitations-skeleton"
      applyHeaderSize
      toolbarProps={{
        searchPlaceholder: t("Search by name, code or ID..."),
        searchDebounceMs: 500,
        filters: [
          {
            columnId: "status",
            title: t("Status"),
            options: invitationStatusOptions,
            singleSelect: true,
          },
        ],
      }}
      mobile={<InvitationsMobileList table={table} isLoading={isLoading} />}
      getRowClassName={(row, { isMobile: isMobileRow }) => {
        if (!isDisabledInvitationRow(row.original)) return undefined;
        return isMobileRow ? DISABLED_ROW_MOBILE : DISABLED_ROW_DESKTOP;
      }}
      bulkActions={<DataTableBulkActions table={table} />}
    />
  );
}
