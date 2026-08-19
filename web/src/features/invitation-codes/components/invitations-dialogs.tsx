import { InvitationsDeleteDialog } from "./invitations-delete-dialog";
import { InvitationsMutateDrawer } from "./invitations-mutate-drawer";
import { useInvitations } from "./invitations-provider";

export function InvitationsDialogs() {
  const { open, setOpen, currentRow, setCurrentRow } = useInvitations();

  return (
    <>
      <InvitationsMutateDrawer
        open={open === "create" || open === "update"}
        onOpenChange={(isOpen) => {
          if (!isOpen) {
            setOpen(null);
            setCurrentRow(null);
          }
        }}
        currentRow={open === "update" ? currentRow || undefined : undefined}
      />
      <InvitationsDeleteDialog
        open={open === "delete"}
        onOpenChange={(isOpen) => {
          if (!isOpen) {
            setOpen(null);
            setCurrentRow(null);
          }
        }}
        currentRow={currentRow}
      />
    </>
  );
}
