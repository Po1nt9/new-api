import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { Row } from "@tanstack/react-table";
import { Copy, Edit, Power, PowerOff, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { DataTableRowActionMenu } from "@/components/data-table/core/row-action-menu";
import { Button } from "@/components/ui/button";
import {
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
} from "@/components/ui/dropdown-menu";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { updateInvitationStatus } from "../api";
import {
  INVITATION_STATUS,
  SUCCESS_MESSAGES,
  ERROR_MESSAGES,
} from "../constants";
import { copyInvitationCode, isInvitationExpired } from "../lib/utils";
import type { Invitation } from "../types";
import { useInvitations } from "./invitations-provider";

interface DataTableRowActionsProps<TData> {
  row: Row<TData>;
}

export function DataTableRowActions<TData>({
  row,
}: DataTableRowActionsProps<TData>) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const invitation = row.original as Invitation;
  const { setOpen, setCurrentRow } = useInvitations();
  const isEnabled = invitation.status === INVITATION_STATUS.ENABLED;
  const isUsed = invitation.status === INVITATION_STATUS.USED;
  const isExpired = isInvitationExpired(
    invitation.expired_time,
    invitation.status
  );

  const toggleStatusMutation = useMutation({
    mutationFn: (newStatus: number) =>
      updateInvitationStatus(invitation.id, newStatus),
    onSuccess: (res) => {
      if (res.success) {
        toast.success(t(SUCCESS_MESSAGES.STATUS_UPDATED));
        queryClient.invalidateQueries({ queryKey: ["invitations"] });
      } else {
        toast.error(res.message || t(ERROR_MESSAGES.UPDATE_FAILED));
      }
    },
    onError: () => {
      toast.error(t(ERROR_MESSAGES.UPDATE_FAILED));
    },
  });

  const handleToggleStatus = () => {
    const newStatus = isEnabled
      ? INVITATION_STATUS.DISABLED
      : INVITATION_STATUS.ENABLED;
    toggleStatusMutation.mutate(newStatus);
  };

  const canEdit = isEnabled && !isExpired;
  const canToggle = !isUsed && !isExpired;

  return (
    <div className="-ml-1.5 flex items-center gap-1">
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => {
                setCurrentRow(invitation);
                setOpen("update");
              }}
              disabled={!canEdit}
              aria-label={t("Edit")}
            />
          }
        >
          <Edit />
        </TooltipTrigger>
        <TooltipContent>{t("Edit")}</TooltipContent>
      </Tooltip>

      <DataTableRowActionMenu ariaLabel={t("Open menu")} modal={false}>
        <DropdownMenuItem onClick={() => copyInvitationCode(invitation.key, t)}>
          <Copy className="mr-2 h-4 w-4" />
          {t("Copy Code")}
        </DropdownMenuItem>

        {canToggle && (
          <DropdownMenuItem onClick={handleToggleStatus}>
            {isEnabled ? (
              <>
                {t("Disable")}
                <DropdownMenuShortcut>
                  <PowerOff size={16} />
                </DropdownMenuShortcut>
              </>
            ) : (
              <>
                {t("Enable")}
                <DropdownMenuShortcut>
                  <Power size={16} />
                </DropdownMenuShortcut>
              </>
            )}
          </DropdownMenuItem>
        )}
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className="text-destructive focus:text-destructive"
          onClick={() => {
            setCurrentRow(invitation);
            setOpen("delete");
          }}
        >
          <Trash2 className="mr-2 h-4 w-4" />
          {t("Delete")}
        </DropdownMenuItem>
      </DataTableRowActionMenu>
    </div>
  );
}
