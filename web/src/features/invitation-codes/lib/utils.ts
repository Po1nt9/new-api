import type { TFunction } from "i18next";
import { toast } from "sonner";
import { INVITATION_STATUS, SUCCESS_MESSAGES } from "../constants";
import type { Invitation } from "../types";

export function copyInvitationCode(code: string, t: TFunction): void {
  navigator.clipboard.writeText(code);
  toast.success(t(SUCCESS_MESSAGES.COPIED));
}

export function isInvitationExpired(
  expiredTime: number,
  status: number
): boolean {
  if (status !== INVITATION_STATUS.ENABLED) return false;
  if (expiredTime === 0) return false;
  return expiredTime < Math.floor(Date.now() / 1000);
}

export function downloadTextFile(
  content: string,
  filename: string,
  mimeType = "text/plain;charset=utf-8"
): void {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

export function exportInvitationsToCSV(items: Invitation[]): void {
  const headers = [
    "ID",
    "Name",
    "Key",
    "Status",
    "Quota",
    "Group",
    "UsedUserID",
    "CreatedTime",
    "ExpiredTime",
    "UsedTime",
  ];
  const rows = items.map((item) => [
    item.id,
    `"${(item.name || "").replaceAll('"', '""')}"`,
    `"${item.key}"`,
    item.status,
    item.quota,
    `"${item.group || ""}"`,
    item.used_user_id || "",
    item.created_time,
    item.expired_time,
    item.used_time,
  ]);
  const csvContent = [headers.join(","), ...rows.map((r) => r.join(","))].join(
    "\n"
  );
  const filename = `invitations_${new Date().toISOString().slice(0, 10)}.csv`;
  downloadTextFile(csvContent, filename, "text/csv;charset=utf-8");
}

export function exportInvitationsToTXT(items: Invitation[]): void {
  const textContent = items
    .map((item) => `${item.name}\t${item.key}`)
    .join("\n");
  const filename = `invitations_${new Date().toISOString().slice(0, 10)}.txt`;
  downloadTextFile(textContent, filename);
}
