import type { ColumnDef } from "@tanstack/react-table";
import { useTranslation } from "react-i18next";
import { MaskedValueDisplay } from "@/components/masked-value-display";
import { StatusBadge } from "@/components/status-badge";
import { TableId } from "@/components/table-id";
import { Checkbox } from "@/components/ui/checkbox";
import { formatQuota, formatTimestampToDate } from "@/lib/format";
import { INVITATION_STATUSES, INVITATION_FILTER_EXPIRED } from "../constants";
import { isInvitationExpired } from "../lib/utils";
import type { Invitation } from "../types";
import { DataTableRowActions } from "./data-table-row-actions";

export function useInvitationsColumns(): ColumnDef<Invitation>[] {
  const { t } = useTranslation();
  return [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          indeterminate={table.getIsSomePageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label={t("Select all")}
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label={t("Select row")}
          className="translate-y-[2px]"
        />
      ),
      enableSorting: false,
      enableHiding: false,
      size: 40,
    },
    {
      accessorKey: "id",
      header: t("ID"),
      meta: { mobileHidden: true },
      cell: ({ row }) => (
        <TableId value={row.getValue("id") as number} className="w-[60px]" />
      ),
      size: 80,
    },
    {
      accessorKey: "name",
      header: t("Name"),
      meta: { mobileTitle: true },
      cell: ({ row }) => (
        <span className="font-medium">{row.getValue("name")}</span>
      ),
      size: 180,
    },
    {
      accessorKey: "status",
      header: t("Status"),
      meta: { mobileBadge: true },
      cell: ({ row }) => {
        const item = row.original;
        const statusValue = row.getValue("status") as number;

        if (isInvitationExpired(item.expired_time, statusValue)) {
          return (
            <StatusBadge
              label={t("Expired")}
              variant="warning"
              copyable={false}
              className="-ml-1.5"
            />
          );
        }

        const config = INVITATION_STATUSES[statusValue];
        if (!config) return null;
        return (
          <StatusBadge
            label={t(config.labelKey)}
            variant={config.variant}
            copyable={false}
            className="-ml-1.5"
          />
        );
      },
      filterFn: (row, id, value) => {
        const item = row.original;
        const statusValue = row.getValue(id) as number;
        if (value.includes(INVITATION_FILTER_EXPIRED)) {
          if (isInvitationExpired(item.expired_time, statusValue)) {
            return true;
          }
        }
        return value.includes(String(statusValue));
      },
      size: 120,
    },
    {
      accessorKey: "quota",
      header: t("Quota"),
      cell: ({ row }) => {
        const quota = row.getValue("quota") as number;
        return (
          <StatusBadge
            label={formatQuota(quota)}
            variant="neutral"
            copyable={false}
            className="-ml-1.5"
          />
        );
      },
      size: 120,
    },
    {
      accessorKey: "group",
      header: t("Group"),
      cell: ({ row }) => {
        const group = row.getValue("group") as string;
        return <span>{group || t("Default")}</span>;
      },
      size: 100,
    },
    {
      id: "code",
      accessorKey: "key",
      header: t("Invitation Code"),
      cell: ({ row }) => {
        const key = (row.getValue("code") as string) || row.original.key;
        const maskedKey =
          key.length > 16
            ? `${key.slice(0, 8)}${"*".repeat(16)}${key.slice(-8)}`
            : key;
        return (
          <MaskedValueDisplay
            label={t("Full Code")}
            fullValue={key}
            maskedValue={maskedKey}
            copyTooltip={t("Copy code")}
            copyAriaLabel={t("Copy invitation code")}
          />
        );
      },
      enableSorting: false,
      size: 300,
    },
    {
      accessorKey: "used_user_id",
      header: t("Used By User ID"),
      cell: ({ row }) => {
        const uid = row.getValue("used_user_id") as number;
        return <span>{uid > 0 ? uid : "-"}</span>;
      },
      size: 120,
    },
    {
      accessorKey: "used_time",
      header: t("Used At"),
      meta: { mobileHidden: true },
      cell: ({ row }) => {
        const usedTime = row.getValue("used_time") as number;
        return (
          <div className="min-w-[160px] font-mono text-sm">
            {usedTime > 0 ? formatTimestampToDate(usedTime) : "-"}
          </div>
        );
      },
      size: 160,
    },
    {
      accessorKey: "created_time",
      header: t("Created"),
      meta: { mobileHidden: true },
      cell: ({ row }) => (
        <div className="min-w-[160px] font-mono text-sm">
          {formatTimestampToDate(row.getValue("created_time") as number)}
        </div>
      ),
      size: 160,
    },
    {
      accessorKey: "expired_time",
      header: t("Expires At"),
      meta: { mobileHidden: true },
      cell: ({ row }) => {
        const expired = row.getValue("expired_time") as number;
        return (
          <div className="min-w-[160px] font-mono text-sm">
            {expired > 0 ? formatTimestampToDate(expired) : t("Never")}
          </div>
        );
      },
      size: 160,
    },
    {
      id: "actions",
      cell: ({ row }) => <DataTableRowActions row={row} />,
      size: 80,
    },
  ];
}
