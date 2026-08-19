# 0002. Empty Response Retry (Streamed and Non-Streamed)

## Status
Accepted (特批后端变更; implemented & deployed 2026-08-19 as v1.0.0-rc.25-po1nt9.3, `RetryOnEmptyResponse=true`)

## Context
On 2026-08-18 14:37–14:52 the OpenCodeZen free upstream (channel 48) intermittently returned
HTTP 200 SSE streams that ended normally (`stream_status: end_reason=done, status=ok`) with zero
output text, zero tool calls and no usage event. New API relayed them as successes, so the built-in
retry machinery (`RetryTimes` + `shouldRetry`) never fired and downstream clients received an empty
turn (`empty_model_response`). Over 24h this produced 45 empty customer responses on `/v1/responses`
plus 3 on `/v1/chat/completions` (see production `logs` table, `type=2 AND completion_tokens=0`).

Upstream tracker state: QuantumNous/new-api#2989 (open, same reproduction), #3275 closed by the
maintainer over double-billing concerns (providers such as Gemini bill even moderation-filtered
empties), PR #2737 abandoned. No official fix exists or is imminent. A production-grade
implementation exists in the AGPL-compatible derivative `unorouter/new-api` and was used as the
design reference.

Two obstacles make a naive fix wrong:
1. Stock stream handlers forward the opener chunk immediately, so by the time an empty stream is
   recognized, bytes are already on the wire and the response cannot fail over cleanly.
2. `controller.Relay`'s error defer writes `c.JSON(...)` unconditionally; an error returned after
   a partially streamed response would append garbage to the SSE stream.

## Decision
1. Buffer stream opener events until the first *meaningful* event arrives (non-empty text delta,
   any `response.output_item.done`, or — on the chat path — any counted content/reasoning/tool
   output). Until then nothing is written to the client, so an empty stream leaves the response
   fully uncommitted.
2. Classify a finished stream as empty when there is no text, no tool invocation and no billed
   completion tokens. An empty stream that never became meaningful returns a retryable
   `empty_response` error with HTTP 502, which flows through the existing retry/failover and
   channel auto-disable machinery (`RetryTimes`, `processChannelError`). An empty stream that
   already streamed meaningful bytes returns the same error with `ErrOptionWithSkipRetry` (logged,
   counted for disable, never replayed to the client).
3. Guard `controller.Relay`'s error defer with `c.Writer.Written()` so post-stream errors never
   corrupt a committed SSE response.
4. Everything is behind a single global option `RetryOnEmptyResponse` (default **false** = exact
   stock behavior). Rationale for the default: upstream rejected this behavior over double-billing
   on paid providers that bill filtered empties; this deployment's channels are free-tier, where
   the trade-off is reversed, so the option is enabled here but the shipped default stays neutral.
5. Content-filter finishes (`finish_reason=content_filter`) are excluded from empty classification:
   moderation outcomes are policy decisions, not transient faults.

Scope note (config, not code): the `/v1/responses` traffic is additionally pinned to a single
channel because only the zen upstream speaks the Responses protocol natively (基元律动/NVIDIA
404 on passthrough). rc.25 ships the `advanced_custom` channel type whose
`openai_responses_to_openai_chat_completions` converter route lets any chat upstream serve
`/v1/responses`; deploying such a channel for the healthy upstream gives Responses traffic real
cross-channel failover. That part is pure configuration and is documented in the ops runbook, not
in this ADR.

## Consequences
- Empty upstream streams become visible, retryable failures instead of silent empty successes.
- First-byte latency is unchanged for healthy streams (buffer releases as soon as meaningful
  output lands; opener events were already subject to one-chunk forwarding delay on the chat path).
- With the option off, byte-for-byte stock behavior is preserved (guarded by regression tests).
- Fork divergence: 4 backend files + tests + option wiring, added to the CI backend-pristine
  allowlist; candidate for an upstream PR behind an opt-in flag.

## References
- QuantumNous/new-api#2989, #3275, PR #2737 (closed, no changes)
- unorouter/new-api `relay/channel/openai/relay-openai.go` (design reference, AGPL-3.0 like new-api)
- Production evidence: HK relay PG `logs` table 2026-08-18 window, request_ids 126197–126242
