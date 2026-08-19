import React, { createContext, useContext, useState } from "react";
import useDialogState from "@/hooks/use-dialog";
import type { Invitation, InvitationsDialogType } from "../types";

type InvitationsContextType = {
  open: InvitationsDialogType | null;
  setOpen: (str: InvitationsDialogType | null) => void;
  currentRow: Invitation | null;
  setCurrentRow: React.Dispatch<React.SetStateAction<Invitation | null>>;
};

const InvitationsContext = createContext<InvitationsContextType | null>(null);

export function InvitationsProvider(props: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<InvitationsDialogType>(null);
  const [currentRow, setCurrentRow] = useState<Invitation | null>(null);

  return (
    <InvitationsContext
      value={{
        open,
        setOpen,
        currentRow,
        setCurrentRow,
      }}
    >
      {props.children}
    </InvitationsContext>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useInvitations() {
  const context = useContext(InvitationsContext);
  if (!context) {
    throw new Error(
      "useInvitations must be used within an InvitationsProvider"
    );
  }
  return context;
}
