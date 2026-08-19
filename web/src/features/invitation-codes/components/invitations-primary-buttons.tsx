import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { deleteInvalidInvitations } from "../api";
import { SUCCESS_MESSAGES, ERROR_MESSAGES } from "../constants";
import { useInvitations } from "./invitations-provider";

export function InvitationsPrimaryButtons() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { setOpen, setCurrentRow } = useInvitations();

  const clearInvalidMutation = useMutation({
    mutationFn: () => deleteInvalidInvitations(),
    onSuccess: (res) => {
      if (res.success) {
        toast.success(t(SUCCESS_MESSAGES.INVALID_CLEARED));
        queryClient.invalidateQueries({ queryKey: ["invitations"] });
      } else {
        toast.error(res.message || t(ERROR_MESSAGES.CLEAR_INVALID_FAILED));
      }
    },
    onError: () => {
      toast.error(t(ERROR_MESSAGES.CLEAR_INVALID_FAILED));
    },
  });

  return (
    <div className="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        disabled={clearInvalidMutation.isPending}
        onClick={() => clearInvalidMutation.mutate()}
      >
        <Trash2 className="mr-1.5 h-4 w-4" />
        {clearInvalidMutation.isPending
          ? t("Clearing...")
          : t("Clear Invalid Codes")}
      </Button>
      <Button
        size="sm"
        onClick={() => {
          setCurrentRow(null);
          setOpen("create");
        }}
      >
        <Plus className="mr-1.5 h-4 w-4" />
        {t("Generate Invitation Codes")}
      </Button>
    </div>
  );
}
