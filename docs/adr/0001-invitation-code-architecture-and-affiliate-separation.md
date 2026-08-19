# 0001. Invitation Code Architecture and Affiliate Separation

## Status
Accepted

## Context
Upstream New API did not have a built-in single-use invitation code system with quota and group provisioning. A historical third-party PR (#6317) stored only SHA-256 hashes without plaintext visibility, lacked initial quota/group bindings, and used overly complex per-channel OAuth method gating. Additionally, the existing platform already used `AffCode` for friend referral commission tracking.

## Decision
We decided to:
1. Replicate the proven, robust design pattern of `Redemption` into `Invitation` (single-use card model, plaintext visibility for admins, batch generation with custom prefixes or explicit custom keys, soft deletes, and CAS atomic redemption).
2. Support initial quota granting and target user group binding directly upon registration.
3. Explicitly decouple `Invitation Code` (access gating / starter grant) from `Affiliate Code` (referrer commission attribution), allowing them to operate independently or concurrently without cross-contamination.
4. Replace complex multi-channel filters with a clean global two-state policy (`InvitationCodeRequired`: true for mandatory gating, false for optional starter grants).

## Consequences
- Clean, consistent admin and user UX aligning directly with existing platform patterns.
- High-concurrency safety via database transaction row locks and atomic CAS status transitions.
- Preserved 100% compatibility with existing affiliate referral links and OAuth flows.
