import { createFileRoute, redirect } from "@tanstack/react-router";
import z from "zod";
import { Invitations } from "@/features/invitation-codes";
import { INVITATION_FILTER_VALUES } from "@/features/invitation-codes/constants";
import { ROLE } from "@/lib/roles";
import { useAuthStore } from "@/stores/auth-store";

const invitationsSearchSchema = z.object({
  page: z.number().optional().catch(1),
  pageSize: z.number().optional().catch(10),
  filter: z.string().optional().catch(""),
  status: z.array(z.enum(INVITATION_FILTER_VALUES)).optional().catch([]),
});

export const Route = createFileRoute("/_authenticated/invitation-codes/")({
  beforeLoad: () => {
    const { auth } = useAuthStore.getState();
    if (!auth.user || auth.user.role < ROLE.ADMIN) {
      throw redirect({
        to: "/403",
      });
    }
  },
  validateSearch: invitationsSearchSchema,
  component: Invitations,
});
