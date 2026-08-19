package openai

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

func OaiResponsesHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)

	// read response body
	var responsesResponse dto.OpenAIResponsesResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeReadResponseBodyFailed, http.StatusInternalServerError)
	}
	err = common.Unmarshal(responseBody, &responsesResponse)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if oaiError := responsesResponse.GetOpenAIError(); oaiError != nil && oaiError.Type != "" {
		return nil, types.WithOpenAIError(*oaiError, resp.StatusCode)
	}

	// (fork, ADR 0002) An empty completed response (no output items, no billed
	// completion) must not be forwarded as a success; fail over instead.
	if operation_setting.RetryOnEmptyResponse && len(responsesResponse.Output) == 0 && (responsesResponse.Usage == nil || responsesResponse.Usage.OutputTokens == 0) {
		logger.LogError(c, "empty non-stream response, retrying on next channel")
		return &dto.Usage{}, emptyResponseError(true)
	}

	// 写入新的 response body
	service.IOCopyBytesGracefully(c, resp, responseBody)

	// compute usage
	usage := dto.Usage{}
	if responsesResponse.Usage != nil {
		usage.PromptTokens = responsesResponse.Usage.InputTokens
		usage.CompletionTokens = responsesResponse.Usage.OutputTokens
		usage.TotalTokens = responsesResponse.Usage.TotalTokens
		if responsesResponse.Usage.InputTokensDetails != nil {
			usage.PromptTokensDetails.CachedTokens = responsesResponse.Usage.InputTokensDetails.CachedTokens
			usage.PromptTokensDetails.CacheWriteTokens = responsesResponse.Usage.InputTokensDetails.CacheWriteTokens
		}
	}
	// Count actual tool invocations from Output (not tool declarations).
	for _, output := range responsesResponse.Output {
		switch output.Type {
		case dto.BuildInCallWebSearchCall:
			info.CountBillableToolCall(dto.BuildInCallWebSearchCall, "")
		case dto.BuildInCallFileSearchCall:
			info.CountBillableToolCall(dto.BuildInCallFileSearchCall, "")
		case dto.BuildInCallFunctionCall:
			info.CountBillableToolCall(dto.BuildInCallFunctionCall, output.Name)
		}
	}

	imageCounter := &relaycommon.ImageGenerationCallCounter{}
	if !relaycommon.IsNonBillableResponsesStatus(responsesResponse.Status) {
		for i := range responsesResponse.Output {
			idx := i
			imageCounter.Observe(&responsesResponse.Output[i], &idx)
		}
	}
	imageCounter.Commit(info)

	return &usage, nil
}

func OaiResponsesStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	if resp == nil || resp.Body == nil {
		logger.LogError(c, "invalid response or response body")
		return nil, types.NewError(fmt.Errorf("invalid response"), types.ErrorCodeBadResponse)
	}

	defer service.CloseResponseBodyGracefully(resp)

	var usage = &dto.Usage{}
	var responseTextBuilder strings.Builder
	imageCounter := &relaycommon.ImageGenerationCallCounter{}
	imageCommitted := false

	// (fork, ADR 0002) Buffer opener events until the first meaningful one so an
	// empty stream can still fail over without having committed bytes to the client.
	buffering := operation_setting.RetryOnEmptyResponse
	var pendingFlush []string
	streamingStarted := false
	sawToolOutput := false

	helper.StreamScannerHandler(c, resp, info, func(data string, sr *helper.StreamResult) {

		// 检查当前数据是否包含 completed 状态和 usage 信息
		var streamResponse dto.ResponsesStreamResponse
		if err := common.UnmarshalJsonStr(data, &streamResponse); err != nil {
			logger.LogError(c, "failed to unmarshal stream response: "+err.Error())
			sr.Error(err)
			return
		}
		if buffering {
			if streamingStarted {
				sendResponsesStreamData(c, streamResponse, data)
			} else if isMeaningfulResponsesEvent(&streamResponse) {
				streamingStarted = true
				for _, pending := range pendingFlush {
					var pendingResponse dto.ResponsesStreamResponse
					if err := common.UnmarshalJsonStr(pending, &pendingResponse); err == nil {
						sendResponsesStreamData(c, pendingResponse, pending)
					}
				}
				pendingFlush = nil
				sendResponsesStreamData(c, streamResponse, data)
			} else {
				pendingFlush = append(pendingFlush, data)
			}
		} else {
			sendResponsesStreamData(c, streamResponse, data)
		}
		switch streamResponse.Type {
		case "response.completed", "response.done":
			if streamResponse.Response != nil {
				if streamResponse.Response.Usage != nil {
					if streamResponse.Response.Usage.InputTokens != 0 {
						usage.PromptTokens = streamResponse.Response.Usage.InputTokens
					}
					if streamResponse.Response.Usage.OutputTokens != 0 {
						usage.CompletionTokens = streamResponse.Response.Usage.OutputTokens
					}
					if streamResponse.Response.Usage.TotalTokens != 0 {
						usage.TotalTokens = streamResponse.Response.Usage.TotalTokens
					}
					if streamResponse.Response.Usage.InputTokensDetails != nil {
						usage.PromptTokensDetails.CachedTokens = streamResponse.Response.Usage.InputTokensDetails.CachedTokens
						usage.PromptTokensDetails.CacheWriteTokens = streamResponse.Response.Usage.InputTokensDetails.CacheWriteTokens
					}
				}
				if !imageCommitted {
					if relaycommon.IsNonBillableResponsesStatus(streamResponse.Response.Status) {
						imageCounter.Reset()
						imageCounter.Commit(info)
						imageCommitted = true
					} else {
						for i := range streamResponse.Response.Output {
							idx := i
							imageCounter.Observe(&streamResponse.Response.Output[i], &idx)
						}
						imageCounter.Commit(info)
						imageCommitted = true
					}
				}
			} else if !imageCommitted {
				imageCounter.Commit(info)
				imageCommitted = true
			}
			if streamResponse.Response != nil && len(streamResponse.Response.Output) > 0 {
				sawToolOutput = true
			}
		case "response.failed", "response.incomplete", "response.cancelled", "response.canceled":
			if !imageCommitted {
				imageCounter.Reset()
				imageCounter.Commit(info)
				imageCommitted = true
			}
		case "response.output_text.delta":
			// 处理输出文本
			responseTextBuilder.WriteString(streamResponse.Delta)
		case dto.ResponsesOutputTypeItemDone:
			if streamResponse.Item != nil {
				switch streamResponse.Item.Type {
				case dto.BuildInCallWebSearchCall:
					info.CountBillableToolCall(dto.BuildInCallWebSearchCall, "")
					sawToolOutput = true
				case dto.BuildInCallFileSearchCall:
					info.CountBillableToolCall(dto.BuildInCallFileSearchCall, "")
					sawToolOutput = true
				case dto.BuildInCallFunctionCall:
					info.CountBillableToolCall(dto.BuildInCallFunctionCall, streamResponse.Item.Name)
					sawToolOutput = true
				case dto.ResponsesOutputTypeImageGenerationCall:
					if !imageCommitted {
						imageCounter.Observe(streamResponse.Item, streamResponse.OutputIndex)
					}
					sawToolOutput = true
				}
			}
		}
	})

	if usage.CompletionTokens == 0 {
		// 计算输出文本的 token 数量
		tempStr := responseTextBuilder.String()
		if len(tempStr) > 0 {
			// 非正常结束，使用输出文本的 token 数量
			completionTokens := service.CountTextToken(tempStr, info.UpstreamModelName)
			usage.CompletionTokens = completionTokens
		}
	}

	if usage.PromptTokens == 0 && usage.CompletionTokens != 0 {
		usage.PromptTokens = info.GetEstimatePromptTokens()
	}

	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

	// (fork, ADR 0002) An empty stream (no text, no tool invocation, no billed
	// completion) must not reach the client as a success. If nothing was ever
	// released from the buffer the response is uncommitted and can be retried on
	// another channel; otherwise the error is logged/counted but never replayed.
	if buffering && responseTextBuilder.Len() == 0 && !sawToolOutput && usage.CompletionTokens == 0 {
		if !streamingStarted {
			logger.LogError(c, "empty response stream, retrying on next channel")
			return usage, emptyResponseError(true)
		}
		return usage, emptyResponseError(false)
	}
	if buffering && !streamingStarted && len(pendingFlush) > 0 {
		// Pathological case: no delta/item events but the stream is not classified
		// empty (e.g. usage-only completion). Release everything that was held.
		for _, pending := range pendingFlush {
			var pendingResponse dto.ResponsesStreamResponse
			if err := common.UnmarshalJsonStr(pending, &pendingResponse); err == nil {
				sendResponsesStreamData(c, pendingResponse, pending)
			}
		}
	}

	return usage, nil
}

// isMeaningfulResponsesEvent reports whether an event carries real model output.
// Declarations (output_item.added), lifecycle events (created/in_progress) and
// reasoning summaries do not count: they must not release the ADR 0002 buffer.
func isMeaningfulResponsesEvent(sr *dto.ResponsesStreamResponse) bool {
	if sr.Type == "response.output_text.delta" {
		return sr.Delta != ""
	}
	if sr.Type == dto.ResponsesOutputTypeItemDone {
		return sr.Item != nil
	}
	return false
}

func emptyResponseError(retryable bool) *types.NewAPIError {
	if retryable {
		return types.NewOpenAIError(
			fmt.Errorf("upstream returned an empty response (no content/tool calls)"),
			types.ErrorCodeEmptyResponse, http.StatusBadGateway,
		)
	}
	return types.NewOpenAIError(
		fmt.Errorf("upstream returned an empty response (no content/tool calls)"),
		types.ErrorCodeEmptyResponse, http.StatusBadGateway,
		types.ErrOptionWithSkipRetry(),
	)
}
