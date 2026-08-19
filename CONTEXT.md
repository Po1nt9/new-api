# New API Gateway & User Onboarding Context

The core API routing, user management, and quota governance context for the AI API gateway.

## Language

### User Onboarding & Access Control

**Invitation Code**:
A single-use administrator-generated credential that gates registration, optionally grants starter quota, or assigns initial user groups.
_Avoid_: Invite link, promo code, referral token, activation key

**Affiliate Code**:
A persistent referral identifier belonging to an existing user that attributes invitee registrations to the referrer for commission and tracking.
_Avoid_: Invitation code, promo code, referral link

**Redemption Code**:
A prepaid voucher used by an existing authenticated account to top up account quota.
_Avoid_: Invitation code, gift card, coupon

**Registration Gating Mode**:
The system-wide policy determining whether an Invitation Code is strictly mandatory or optional during user account creation.
_Avoid_: Invite-only toggle, whitelist mode

### Quota & Billing

**Quota**:
The universal internal unit of computational credit consumed across AI provider channels.
_Avoid_: Tokens, credits, balance, points

**User Group**:
A classification tier assigned to users that determines routing priority, available channel bindings, and rate limits.
_Avoid_: Role, level, permission group
