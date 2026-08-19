import type { Table as TanstackTable } from "@tanstack/react-table";
import { useTranslation } from "react-i18next";
import { MaskedValueDisplay } from "@/components/masked-value-display";
import { StatusBadge } from "@/components/status-badge";
import { Skeleton } from "@/components/ui/skeleton";
import { formatQuota, formatTimestampToDate } from "@/lib/format";
import { INVITATION_STATUSES } from "../constants";
import { isInvitationExpired } from "../lib/utils";
import type { Invitation } from "../types";
import { DataTableRowActions } from "./data-table-row-actions";

interface InvitationsMobileListProps {
  table: TanstackTable<Invitation>;
  isLoading?: boolean;
}

export function InvitationsMobileList(props: InvitationsMobileListProps) {
  const { t } = useTranslation();
  const rows = props.table.getRowModel().rows;

  if (props.isLoading) {
    return (
      <div className="space-y-3">
        {[1, 2, 3].map((i) => (
          <div key={i} className="rounded-lg border p-4 space-y-2">
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-4 w-48" />
          </div>
        ))}
      </div>
    );
  }

  if (rows.length === 0) {
    return (
      <div className="py-8 text-center text-muted-foreground text-sm">
        {t("No invitation codes found")}
      </div>
    );
  }

  return (
    <div className="divide-y rounded-lg border">
      {rows.map((row) => {
        const item = row.original;
        const isExp = isInvitationExpired(item.expired_time, item.status);
        const config = INVITATION_STATUSES[item.status];
        const maskedKey =
          item.key.length > 16
            ? `${item.key.slice(0, 8)}${"*".repeat(16)}${item.key.slice(-8)}`
            : item.key;

        let badge = null;
        if (isExp) {
          badge = (
            <StatusBadge
              label={t("Expired")}
              variant="warning"
              copyable={false}
            />
          );
        } else if (config) {
          badge = (
            <StatusBadge
              label={t(config.labelKey)}
              variant={config.variant}
              copyable={false}
            />
          );
        }

        return (
          <div key={item.id} className="p-3.5 space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-medium text-sm">{item.name}</span>
              <div className="flex items-center gap-2">
                {badge}
                <DataTableRowActions row={row} />
              </div>
            </div>
            <div className="text-xs text-muted-foreground flex items-center justify-between">
              <span>
                {t("Quota")}: {formatQuota(item.quota)}
              </span>
              <span>
                {t("Group")}: {item.group || t("Default")}
              </span>
            </div>
            <div className="text-xs flex items-center gap-1.5">
              <span className="text-muted-foreground">{t("Code")}:</span>
              <MaskedValueDisplay
                label={t("Full Code")}
                fullValue={item.key}
                maskedValue={maskedKey}
                copyTooltip={t("Copy code")}
                copyAriaLabel={t("Copy invitation code")}
              />
            </div>
            {item.used_user_id > 0 && (
              <div className="text-xs text-muted-foreground flex justify-between bg-muted/30 p-1.5 rounded">
                <span>
                  {t("Used By User ID")}: {item.used_user_id}
                </span>
                <span>
                  {item.used_time > 0
                    ? formatTimestampToDate(item.used_time)
                    : "-"}
                </span>
              </div>
            )}
            <div className="text-xs text-muted-foreground flex justify-between pt-1">
              <span>{formatTimestampToDate(item.created_time)}</span>
              <span>
                {item.expired_time > 0
                  ? formatTimestampToDate(item.expired_time)
                  : t("Never")}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}
