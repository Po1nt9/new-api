import type { TFunction } from "i18next";
import type { StatusBadgeProps } from "@/components/status-badge";

export const INVITATION_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
  USED: 3,
} as const;

export const INVITATION_STATUS_VALUES = Object.values(INVITATION_STATUS).map(
  (value) => String(value)
) as `${number}`[];

export const INVITATION_STATUSES: Record<
  number,
  Pick<StatusBadgeProps, "variant"> & {
    labelKey: string;
    value: number;
  }
> = {
  [INVITATION_STATUS.ENABLED]: {
    labelKey: "Unused",
    variant: "success",
    value: INVITATION_STATUS.ENABLED,
  },
  [INVITATION_STATUS.DISABLED]: {
    labelKey: "Disabled",
    variant: "neutral",
    value: INVITATION_STATUS.DISABLED,
  },
  [INVITATION_STATUS.USED]: {
    labelKey: "Used",
    variant: "neutral",
    value: INVITATION_STATUS.USED,
  },
} as const;

export const INVITATION_FILTER_EXPIRED = "expired";

export const INVITATION_FILTER_VALUES = [
  String(INVITATION_STATUS.ENABLED),
  String(INVITATION_STATUS.DISABLED),
  String(INVITATION_STATUS.USED),
  INVITATION_FILTER_EXPIRED,
] as const;

export function getInvitationStatusOptions(t: TFunction) {
  return [
    ...Object.values(INVITATION_STATUSES).map((config) => ({
      label: t(config.labelKey),
      value: String(config.value),
    })),
    {
      label: t("Expired"),
      value: INVITATION_FILTER_EXPIRED,
    },
  ];
}

export const INVITATION_VALIDATION = {
  NAME_MIN_LENGTH: 1,
  NAME_MAX_LENGTH: 20,
  COUNT_MIN: 1,
  COUNT_MAX: 100,
} as const;

export function getInvitationFormErrorMessages(t: TFunction) {
  return {
    NAME_REQUIRED: t("Please enter a name for the invitation code"),
    NAME_LENGTH_INVALID: t(
      "Invitation code name length must be between {{min}}-{{max}} characters",
      {
        min: INVITATION_VALIDATION.NAME_MIN_LENGTH,
        max: INVITATION_VALIDATION.NAME_MAX_LENGTH,
      }
    ),
    COUNT_INVALID: t("Count must be between {{min}} and {{max}}", {
      min: INVITATION_VALIDATION.COUNT_MIN,
      max: INVITATION_VALIDATION.COUNT_MAX,
    }),
  };
}

export const SUCCESS_MESSAGES = {
  CREATED: "Invitation code(s) created successfully",
  UPDATED: "Invitation code updated successfully",
  DELETED: "Invitation code deleted successfully",
  COPIED: "Invitation code copied to clipboard",
  STATUS_UPDATED: "Status updated successfully",
  BATCH_DELETED: "Selected invitation codes deleted",
  INVALID_CLEARED: "Invalid invitation codes cleared",
} as const;

export const ERROR_MESSAGES = {
  CREATE_FAILED: "Failed to save invitation code",
  UPDATE_FAILED: "Failed to update invitation code",
  DELETE_FAILED: "Failed to delete invitation code",
  BATCH_DELETE_FAILED: "Failed to delete selected invitation codes",
  CLEAR_INVALID_FAILED: "Failed to delete invalid invitation codes",
  LOAD_FAILED: "Failed to load invitation codes",
  SEARCH_FAILED: "Failed to search invitation codes",
} as const;
