import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { Table } from "@tanstack/react-table";
import { Download, FileText, Trash2 } from "lucide-react";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { CopyButton } from "@/components/copy-button";
import { DataTableBulkActions as BulkActionsToolbar } from "@/components/data-table";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { batchDeleteInvitations } from "../api";
import { SUCCESS_MESSAGES, ERROR_MESSAGES } from "../constants";
import { exportInvitationsToCSV, exportInvitationsToTXT } from "../lib/utils";
import type { Invitation } from "../types";

type DataTableBulkActionsProps<TData> = {
  table: Table<TData>;
};

export function DataTableBulkActions<TData>(
  props: DataTableBulkActionsProps<TData>
) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const selectedRows = props.table.getSelectedRowModel().rows;

  const selectedInvitations = useMemo(() => {
    return selectedRows.map((row) => row.original as Invitation);
  }, [selectedRows]);

  const contentToCopy = useMemo(() => {
    return selectedInvitations
      .map((inv) => `${inv.name}\t${inv.key}`)
      .join("\n");
  }, [selectedInvitations]);

  const batchDeleteMutation = useMutation({
    mutationFn: (ids: number[]) => batchDeleteInvitations(ids),
    onSuccess: (res) => {
      if (res.success) {
        toast.success(t(SUCCESS_MESSAGES.BATCH_DELETED));
        props.table.resetRowSelection();
        queryClient.invalidateQueries({ queryKey: ["invitations"] });
      } else {
        toast.error(res.message || t(ERROR_MESSAGES.BATCH_DELETE_FAILED));
      }
    },
    onError: () => {
      toast.error(t(ERROR_MESSAGES.BATCH_DELETE_FAILED));
    },
  });

  const handleExportCSV = () => {
    if (selectedInvitations.length === 0) return;
    exportInvitationsToCSV(selectedInvitations);
  };

  const handleExportTXT = () => {
    if (selectedInvitations.length === 0) return;
    exportInvitationsToTXT(selectedInvitations);
  };

  const handleBatchDelete = () => {
    const ids = selectedInvitations.map((inv) => inv.id);
    if (ids.length === 0) return;
    batchDeleteMutation.mutate(ids);
  };

  return (
    <BulkActionsToolbar table={props.table} entityName={t("Invitation Codes")}>
      <CopyButton
        value={contentToCopy}
        variant="outline"
        size="icon"
        className="size-8"
        tooltip={t("Copy selected codes")}
        successTooltip={t("Codes copied!")}
        aria-label={t("Copy selected codes")}
      />

      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant="outline"
              size="icon"
              className="size-8"
              onClick={handleExportCSV}
              aria-label={t("Export to CSV")}
            />
          }
        >
          <Download className="h-4 w-4" />
        </TooltipTrigger>
        <TooltipContent>{t("Export to CSV")}</TooltipContent>
      </Tooltip>

      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant="outline"
              size="icon"
              className="size-8"
              onClick={handleExportTXT}
              aria-label={t("Export to TXT")}
            />
          }
        >
          <FileText className="h-4 w-4" />
        </TooltipTrigger>
        <TooltipContent>{t("Export to TXT")}</TooltipContent>
      </Tooltip>

      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant="outline"
              size="icon"
              className="size-8 text-destructive hover:bg-destructive hover:text-destructive-foreground"
              disabled={batchDeleteMutation.isPending}
              onClick={handleBatchDelete}
              aria-label={t("Delete selected")}
            />
          }
        >
          <Trash2 className="h-4 w-4" />
        </TooltipTrigger>
        <TooltipContent>{t("Delete selected")}</TooltipContent>
      </Tooltip>
    </BulkActionsToolbar>
  );
}
