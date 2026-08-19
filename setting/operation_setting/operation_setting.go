package operation_setting

import "strings"

var DemoSiteEnabled = false
var SelfUseModeEnabled = false

// RetryOnEmptyResponse (fork, ADR 0002): when true, a stream/non-stream response that
// finishes with no text, no tool calls and no billed completion tokens is converted
// into a retryable 502 empty_response error instead of being relayed as a success.
// Default false = upstream stock behavior; upstream rejected auto-retry over
// double-billing concerns on providers that bill moderation-filtered empties.
var RetryOnEmptyResponse = false

var AutomaticDisableKeywords = []string{
	"Your credit balance is too low",
	"This organization has been disabled.",
	"You exceeded your current quota",
	"Permission denied",
	"The security token included in the request is invalid",
	"Operation not allowed",
	"Your account is not authorized",
}

func AutomaticDisableKeywordsToString() string {
	return strings.Join(AutomaticDisableKeywords, "\n")
}

func AutomaticDisableKeywordsFromString(s string) {
	AutomaticDisableKeywords = []string{}
	ak := strings.Split(s, "\n")
	for _, k := range ak {
		k = strings.TrimSpace(k)
		k = strings.ToLower(k)
		if k != "" {
			AutomaticDisableKeywords = append(AutomaticDisableKeywords, k)
		}
	}
}
