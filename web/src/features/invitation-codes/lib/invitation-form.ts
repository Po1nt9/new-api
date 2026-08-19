import type { TFunction } from "i18next";
import { z } from "zod";
import {
  parseQuotaFromDollars,
  quotaUnitsToEditableAmount,
} from "@/lib/format";
import {
  INVITATION_VALIDATION,
  getInvitationFormErrorMessages,
} from "../constants";
import type { InvitationFormData, Invitation } from "../types";

export function getInvitationFormSchema(t: TFunction) {
  const msg = getInvitationFormErrorMessages(t);
  return z.object({
    name: z
      .string()
      .min(INVITATION_VALIDATION.NAME_MIN_LENGTH, msg.NAME_LENGTH_INVALID)
      .max(INVITATION_VALIDATION.NAME_MAX_LENGTH, msg.NAME_LENGTH_INVALID),
    prefix: z.string().optional(),
    key: z.string().optional(),
    quota_dollars: z.number().min(0, t("Quota must be a positive number")),
    group: z.string().optional(),
    expired_time: z.date().optional(),
    count: z
      .number()
      .min(INVITATION_VALIDATION.COUNT_MIN, msg.COUNT_INVALID)
      .max(INVITATION_VALIDATION.COUNT_MAX, msg.COUNT_INVALID)
      .optional(),
  });
}

export type InvitationFormValues = {
  name: string;
  prefix?: string;
  key?: string;
  quota_dollars: number;
  group?: string;
  expired_time?: Date;
  count?: number;
};

export const INVITATION_FORM_DEFAULT_VALUES: InvitationFormValues = {
  name: "",
  prefix: "",
  key: "",
  quota_dollars: 0,
  group: "",
  expired_time: undefined,
  count: 1,
};

export function transformFormDataToPayload(
  data: InvitationFormValues
): InvitationFormData {
  return {
    name: data.name,
    prefix: data.prefix,
    key: data.key,
    quota: parseQuotaFromDollars(data.quota_dollars),
    group: data.group || "",
    expired_time: data.expired_time
      ? Math.floor(data.expired_time.getTime() / 1000)
      : 0,
    count: data.count || 1,
  };
}

export function transformInvitationToFormDefaults(
  invitation: Invitation
): InvitationFormValues {
  return {
    name: invitation.name,
    quota_dollars: quotaUnitsToEditableAmount(invitation.quota),
    group: invitation.group || "",
    expired_time:
      invitation.expired_time > 0
        ? new Date(invitation.expired_time * 1000)
        : undefined,
    count: 1,
  };
}
