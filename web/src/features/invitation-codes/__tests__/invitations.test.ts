import { describe, it, expect } from "vitest";
import { INVITATION_STATUS, INVITATION_STATUSES } from "../constants";
import {
  isInvitationExpired,
  transformFormDataToPayload,
  transformInvitationToFormDefaults,
} from "../lib";
import type { Invitation } from "../types";

describe("Invitation Codes Feature Utilities", () => {
  it("correctly identifies expired invitation codes", () => {
    const now = Math.floor(Date.now() / 1000);

    // Not expired: expired_time is 0 (never)
    expect(isInvitationExpired(0, INVITATION_STATUS.ENABLED)).toBe(false);

    // Not expired: future timestamp
    expect(isInvitationExpired(now + 3600, INVITATION_STATUS.ENABLED)).toBe(
      false
    );

    // Expired: past timestamp and enabled
    expect(isInvitationExpired(now - 100, INVITATION_STATUS.ENABLED)).toBe(
      true
    );

    // Not expired if already used or disabled
    expect(isInvitationExpired(now - 100, INVITATION_STATUS.USED)).toBe(false);
    expect(isInvitationExpired(now - 100, INVITATION_STATUS.DISABLED)).toBe(
      false
    );
  });

  it("transforms form data to API payload correctly", () => {
    const values = {
      name: "Campaign A",
      prefix: "TEST-",
      key: "CUSTOM888",
      quota_dollars: 5,
      group: "vip",
      expired_time: new Date(1700000000 * 1000),
      count: 10,
    };

    const payload = transformFormDataToPayload(values);

    expect(payload.name).toBe("Campaign A");
    expect(payload.prefix).toBe("TEST-");
    expect(payload.key).toBe("CUSTOM888");
    expect(payload.quota).toBe(5 * 500000);
    expect(payload.group).toBe("vip");
    expect(payload.expired_time).toBe(1700000000);
    expect(payload.count).toBe(10);
  });

  it("transforms invitation entity to form default values", () => {
    const invitation: Invitation = {
      id: 1,
      user_id: 1,
      key: "INV-123456",
      status: INVITATION_STATUS.ENABLED,
      name: "VIP User",
      quota: 5000000,
      group: "vip",
      created_time: 1690000000,
      used_time: 0,
      used_user_id: 0,
      expired_time: 1700000000,
    };

    const formValues = transformInvitationToFormDefaults(invitation);

    expect(formValues.name).toBe("VIP User");
    expect(formValues.quota_dollars).toBe(10);
    expect(formValues.group).toBe("vip");
    expect(formValues.expired_time?.getTime()).toBe(1700000000 * 1000);
    expect(formValues.count).toBe(1);
  });

  it("has complete status badge mapping", () => {
    expect(INVITATION_STATUSES[INVITATION_STATUS.ENABLED]).toBeDefined();
    expect(INVITATION_STATUSES[INVITATION_STATUS.DISABLED]).toBeDefined();
    expect(INVITATION_STATUSES[INVITATION_STATUS.USED]).toBeDefined();
  });
});
