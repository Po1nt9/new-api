package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setRetryOnEmptyForTest(t *testing.T, enabled bool) {
	t.Helper()
	old := operation_setting.RetryOnEmptyResponse
	operation_setting.RetryOnEmptyResponse = enabled
	t.Cleanup(func() {
		operation_setting.RetryOnEmptyResponse = old
	})
}

func newStreamContextAndInfo(t *testing.T, path string, relayMode int, relayFormat types.RelayFormat) (*httptest.ResponseRecorder, *gin.Context, *relaycommon.RelayInfo) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() {
		constant.StreamingTimeout = oldTimeout
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, path, nil)
	c.Set(common.RequestIdKey, "empty-response-retry-test")
	info := &relaycommon.RelayInfo{
		OriginModelName: "m",
		RelayMode:       relayMode,
		RelayFormat:     relayFormat,
		DisablePing:     true,
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "m",
		},
	}
	return w, c, info
}

func sseBody(events ...string) io.Reader {
	var b strings.Builder
	for _, e := range events {
		b.WriteString("data: ")
		b.WriteString(e)
		b.WriteString("\n\n")
	}
	b.WriteString("data: [DONE]\n\n")
	return strings.NewReader(b.String())
}

func newSSEResponse(r io.Reader) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(r),
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
	}
}

// --- /v1/responses stream ---

func TestResponsesStreamEmptyRetriesWhenEnabled(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	w, c, info := newStreamContextAndInfo(t, "/v1/responses", relayconstant.RelayModeResponses, types.RelayFormatOpenAIResponses)

	resp := newSSEResponse(sseBody(
		`{"type":"response.created","response":{"id":"r1"}}`,
		`{"type":"response.completed","response":{"id":"r1","status":"completed","output":[]}}`,
	))

	usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.NotNil(t, apiErr)
	assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	assert.Equal(t, types.ErrorCodeEmptyResponse, apiErr.GetErrorCode())
	assert.False(t, types.IsSkipRetryError(apiErr), "empty unstarted stream must stay retryable")
	assert.NotNil(t, usage)
	assert.Equal(t, 0, w.Body.Len(), "nothing may be written to the client for an empty stream")
}

func TestResponsesStreamFailedEventRetriesWhenEnabled(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	w, c, info := newStreamContextAndInfo(t, "/v1/responses", relayconstant.RelayModeResponses, types.RelayFormatOpenAIResponses)

	resp := newSSEResponse(sseBody(
		`{"type":"response.created","response":{"id":"r1"}}`,
		`{"type":"response.failed","response":{"id":"r1","status":"failed"}}`,
	))

	_, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.NotNil(t, apiErr)
	assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	assert.Equal(t, 0, w.Body.Len())
}

func TestResponsesStreamHealthyStillStreamsWhenEnabled(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	w, c, info := newStreamContextAndInfo(t, "/v1/responses", relayconstant.RelayModeResponses, types.RelayFormatOpenAIResponses)

	resp := newSSEResponse(sseBody(
		`{"type":"response.created","response":{"id":"r1"}}`,
		`{"type":"response.output_text.delta","delta":"你好"}`,
		`{"type":"response.completed","response":{"id":"r1","status":"completed","usage":{"input_tokens":5,"output_tokens":2,"total_tokens":7}}}`,
	))

	usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 2, usage.CompletionTokens)
	assert.Contains(t, w.Body.String(), "你好")
	assert.Contains(t, w.Body.String(), "response.created", "buffered openers must be released on first meaningful event")
	assert.Contains(t, w.Body.String(), "response.completed")
}

func TestResponsesStreamEmptyPassthroughWhenDisabled(t *testing.T) {
	setRetryOnEmptyForTest(t, false)
	w, c, info := newStreamContextAndInfo(t, "/v1/responses", relayconstant.RelayModeResponses, types.RelayFormatOpenAIResponses)

	resp := newSSEResponse(sseBody(
		`{"type":"response.created","response":{"id":"r1"}}`,
		`{"type":"response.completed","response":{"id":"r1","status":"completed","output":[]}}`,
	))

	usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr, "option off must keep stock behavior")
	require.NotNil(t, usage)
	assert.Contains(t, w.Body.String(), "response.completed")
}

// --- /v1/responses non-stream ---

func TestResponsesNonStreamEmptyRetriesWhenEnabled(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	w, c, info := newStreamContextAndInfo(t, "/v1/responses", relayconstant.RelayModeResponses, types.RelayFormatOpenAIResponses)

	body, err := common.Marshal(dto.OpenAIResponsesResponse{Status: json.RawMessage(`"completed"`), Output: []dto.ResponsesOutput{}})
	require.NoError(t, err)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	usage, apiErr := OaiResponsesHandler(c, info, resp)
	require.NotNil(t, apiErr)
	assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	assert.Equal(t, types.ErrorCodeEmptyResponse, apiErr.GetErrorCode())
	assert.False(t, types.IsSkipRetryError(apiErr))
	require.NotNil(t, usage)
	assert.Equal(t, 0, w.Body.Len(), "empty non-stream response must not be forwarded")
}

func TestResponsesNonStreamHealthyForwardsWhenEnabled(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	w, c, info := newStreamContextAndInfo(t, "/v1/responses", relayconstant.RelayModeResponses, types.RelayFormatOpenAIResponses)

	body, err := common.Marshal(dto.OpenAIResponsesResponse{
		Status: json.RawMessage(`"completed"`),
		Output: []dto.ResponsesOutput{{Type: "message"}},
		Usage:  &dto.Usage{InputTokens: 3, OutputTokens: 4, TotalTokens: 7},
	})
	require.NoError(t, err)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	usage, apiErr := OaiResponsesHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 4, usage.CompletionTokens)
	assert.True(t, w.Body.Len() > 0)
}

// --- /v1/chat/completions stream ---

func TestChatStreamEmptyRetriesWhenEnabled(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	w, c, info := newStreamContextAndInfo(t, "/v1/chat/completions", relayconstant.RelayModeChatCompletions, types.RelayFormatOpenAI)

	resp := newSSEResponse(sseBody(
		`{"id":"c1","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`,
		`{"id":"c1","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
	))

	usage, apiErr := OaiStreamHandler(c, info, resp)
	require.NotNil(t, apiErr)
	assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	assert.Equal(t, types.ErrorCodeEmptyResponse, apiErr.GetErrorCode())
	assert.False(t, types.IsSkipRetryError(apiErr))
	require.NotNil(t, usage)
	assert.Equal(t, 0, w.Body.Len(), "empty stream must leave the client response uncommitted")
}

func TestChatStreamHealthyStillStreamsWhenEnabled(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	w, c, info := newStreamContextAndInfo(t, "/v1/chat/completions", relayconstant.RelayModeChatCompletions, types.RelayFormatOpenAI)

	resp := newSSEResponse(sseBody(
		`{"id":"c1","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`,
		`{"id":"c1","choices":[{"index":0,"delta":{"content":"答案"},"finish_reason":null}]}`,
		`{"id":"c1","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
	))

	usage, apiErr := OaiStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Contains(t, w.Body.String(), "答案")
	assert.Contains(t, w.Body.String(), "\"role\":\"assistant\"", "buffered opener must be released")
}

func TestChatStreamReasoningOnlyStillForwardsWhenEnabled(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	w, c, info := newStreamContextAndInfo(t, "/v1/chat/completions", relayconstant.RelayModeChatCompletions, types.RelayFormatOpenAI)

	resp := newSSEResponse(sseBody(
		`{"id":"c1","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`,
		`{"id":"c1","choices":[{"index":0,"delta":{"reasoning_content":"思考中"},"finish_reason":null}]}`,
		`{"id":"c1","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
	))

	_, apiErr := OaiStreamHandler(c, info, resp)
	require.Nil(t, apiErr, "reasoning output counts as meaningful; must not be retried")
	assert.Contains(t, w.Body.String(), "思考中")
}

func TestChatStreamContentFilterNotClassifiedEmpty(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	_, c, info := newStreamContextAndInfo(t, "/v1/chat/completions", relayconstant.RelayModeChatCompletions, types.RelayFormatOpenAI)

	resp := newSSEResponse(sseBody(
		`{"id":"c1","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`,
		`{"id":"c1","choices":[{"index":0,"delta":{},"finish_reason":"content_filter"}]}`,
	))

	_, apiErr := OaiStreamHandler(c, info, resp)
	require.Nil(t, apiErr, "content_filter finish is a policy outcome, not an empty response")
}

func TestChatStreamEmptyPassthroughWhenDisabled(t *testing.T) {
	setRetryOnEmptyForTest(t, false)
	w, c, info := newStreamContextAndInfo(t, "/v1/chat/completions", relayconstant.RelayModeChatCompletions, types.RelayFormatOpenAI)

	resp := newSSEResponse(sseBody(
		`{"id":"c1","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`,
		`{"id":"c1","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
	))

	usage, apiErr := OaiStreamHandler(c, info, resp)
	require.Nil(t, apiErr, "option off must keep stock behavior")
	require.NotNil(t, usage)
	assert.True(t, w.Body.Len() > 0)
}

// --- /v1/chat/completions non-stream ---

func TestChatNonStreamEmptyRetriesWhenEnabled(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	w, c, info := newStreamContextAndInfo(t, "/v1/chat/completions", relayconstant.RelayModeChatCompletions, types.RelayFormatOpenAI)

	body := []byte(`{"id":"c1","choices":[{"index":0,"message":{"role":"assistant","content":""},"finish_reason":"stop"}],"usage":{"prompt_tokens":9,"completion_tokens":0,"total_tokens":9}}`)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	usage, apiErr := OpenaiHandler(c, info, resp)
	require.NotNil(t, apiErr)
	assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	assert.Equal(t, types.ErrorCodeEmptyResponse, apiErr.GetErrorCode())
	assert.False(t, types.IsSkipRetryError(apiErr))
	require.NotNil(t, usage)
	assert.Equal(t, 0, w.Body.Len(), "empty non-stream response must not be forwarded")
}

func TestChatNonStreamHealthyForwardsWhenEnabled(t *testing.T) {
	setRetryOnEmptyForTest(t, true)
	w, c, info := newStreamContextAndInfo(t, "/v1/chat/completions", relayconstant.RelayModeChatCompletions, types.RelayFormatOpenAI)

	body := []byte(`{"id":"c1","choices":[{"index":0,"message":{"role":"assistant","content":"好的"},"finish_reason":"stop"}],"usage":{"prompt_tokens":9,"completion_tokens":3,"total_tokens":12}}`)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	usage, apiErr := OpenaiHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 3, usage.CompletionTokens)
	assert.Contains(t, w.Body.String(), "好的")
}
