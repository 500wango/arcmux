package gemini

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	rootconstant "github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/relayconvert"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/model_setting"
	"github.com/QuantumNous/new-api/setting/reasoning"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

type Adaptor struct {
}

func (a *Adaptor) ConvertGeminiRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeminiChatRequest) (any, error) {
	if len(request.Contents) > 0 {
		for i, content := range request.Contents {
			if i == 0 {
				if request.Contents[0].Role == "" {
					request.Contents[0].Role = "user"
				}
			}
			for _, part := range content.Parts {
				if part.FileData != nil {
					if part.FileData.MimeType == "" && strings.Contains(part.FileData.FileUri, "www.youtube.com") {
						part.FileData.MimeType = "video/webm"
					}
				}
			}
		}
	}
	return request, nil
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, req *dto.ClaudeRequest) (any, error) {
	result, err := relayconvert.ConvertRequest(c, info, types.RelayFormatGemini, req)
	if err != nil {
		return nil, err
	}
	geminiRequest, ok := result.Value.(*dto.GeminiChatRequest)
	if !ok {
		return nil, fmt.Errorf("expected Gemini generateContent request, got %T", result.Value)
	}
	return geminiRequest, nil
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	if !strings.HasPrefix(info.UpstreamModelName, "imagen") {
		return nil, errors.New("not supported model for image generation, only imagen models are supported")
	}

	// convert size to aspect ratio but allow user to specify aspect ratio
	aspectRatio := "1:1" // default aspect ratio
	size := strings.TrimSpace(request.Size)
	if size != "" {
		if strings.Contains(size, ":") {
			aspectRatio = size
		} else {
			switch size {
			case "256x256", "512x512", "1024x1024":
				aspectRatio = "1:1"
			case "1536x1024":
				aspectRatio = "3:2"
			case "1024x1536":
				aspectRatio = "2:3"
			case "1024x1792":
				aspectRatio = "9:16"
			case "1792x1024":
				aspectRatio = "16:9"
			}
		}
	}

	// build gemini imagen request
	geminiRequest := dto.GeminiImageRequest{
		Instances: []dto.GeminiImageInstance{
			{
				Prompt: request.Prompt,
			},
		},
		Parameters: dto.GeminiImageParameters{
			SampleCount:      int(lo.FromPtrOr(request.N, uint(1))),
			AspectRatio:      aspectRatio,
			PersonGeneration: "allow_adult", // default allow adult
		},
	}

	// Set imageSize when quality parameter is specified
	// Map quality parameter to imageSize (only supported by Standard and Ultra models)
	// quality values: auto, high, medium, low (for gpt-image-1), hd, standard (for dall-e-3)
	// imageSize values: 1K (default), 2K
	// https://ai.google.dev/gemini-api/docs/imagen
	// https://platform.openai.com/docs/api-reference/images/create
	if request.Quality != "" {
		imageSize := "1K" // default
		switch request.Quality {
		case "hd", "high":
			imageSize = "2K"
		case "2K":
			imageSize = "2K"
		case "standard", "medium", "low", "auto", "1K":
			imageSize = "1K"
		default:
			// unknown quality value, default to 1K
			imageSize = "1K"
		}
		geminiRequest.Parameters.ImageSize = imageSize
	}

	return geminiRequest, nil
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {

}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	oauthProvider := info.UpstreamOAuthProvider
	if oauthProvider == "gemini-cli" || oauthProvider == "antigravity" {
		action := "generateContent"
		if info.IsStream {
			action = "streamGenerateContent"
		}
		return fmt.Sprintf("https://cloudcode-pa.googleapis.com/v1internal:%s", action), nil
	}

	if model_setting.GetGeminiSettings().ThinkingAdapterEnabled &&
		!model_setting.ShouldPreserveThinkingSuffix(info.OriginModelName) {
		// 新增逻辑：处理 -thinking-<budget> 格式
		if strings.Contains(info.UpstreamModelName, "-thinking-") {
			parts := strings.Split(info.UpstreamModelName, "-thinking-")
			info.UpstreamModelName = parts[0]
		} else if strings.HasSuffix(info.UpstreamModelName, "-thinking") { // 旧的适配
			info.UpstreamModelName = strings.TrimSuffix(info.UpstreamModelName, "-thinking")
		} else if strings.HasSuffix(info.UpstreamModelName, "-nothinking") {
			info.UpstreamModelName = strings.TrimSuffix(info.UpstreamModelName, "-nothinking")
		} else if baseModel, level, ok := reasoning.TrimEffortSuffix(info.UpstreamModelName); ok && level != "" {
			info.UpstreamModelName = baseModel
		}
	}

	version := model_setting.GetGeminiVersionSetting(info.UpstreamModelName)

	if strings.HasPrefix(info.UpstreamModelName, "imagen") {
		return fmt.Sprintf("%s/%s/models/%s:predict", info.ChannelBaseUrl, version, info.UpstreamModelName), nil
	}

	if strings.HasPrefix(info.UpstreamModelName, "text-embedding") ||
		strings.HasPrefix(info.UpstreamModelName, "embedding") ||
		strings.HasPrefix(info.UpstreamModelName, "gemini-embedding") {
		action := "embedContent"
		if info.IsGeminiBatchEmbedding {
			action = "batchEmbedContents"
		}
		return fmt.Sprintf("%s/%s/models/%s:%s", info.ChannelBaseUrl, version, info.UpstreamModelName, action), nil
	}

	action := "generateContent"
	if info.IsStream {
		action = "streamGenerateContent?alt=sse"
		if info.RelayMode == constant.RelayModeGemini {
			info.DisablePing = true
		}
	}
	return fmt.Sprintf("%s/%s/models/%s:%s", info.ChannelBaseUrl, version, info.UpstreamModelName, action), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	oauthProvider := common.GetContextKeyString(c, rootconstant.ContextKeyUpstreamOAuthProvider)
	if oauthProvider == "gemini-cli" || oauthProvider == "antigravity" {
		req.Del("x-goog-api-key")
		req.Set("Authorization", "Bearer "+info.ApiKey)
		req.Set("Content-Type", "application/json")
		req.Set("User-Agent", "gemini-cli/0.1")
		return nil
	}
	req.Set("x-goog-api-key", info.ApiKey)
	return nil
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	result, err := relayconvert.ConvertRequest(c, info, types.RelayFormatGemini, request)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	if request.Input == nil {
		return nil, errors.New("input is required")
	}

	inputs := request.ParseInput()
	if len(inputs) == 0 {
		return nil, errors.New("input is empty")
	}
	// We always build a batch-style payload with `requests`, so ensure we call the
	// batch endpoint upstream to avoid payload/endpoint mismatches.
	info.IsGeminiBatchEmbedding = true
	// process all inputs
	geminiRequests := make([]map[string]interface{}, 0, len(inputs))
	for _, input := range inputs {
		geminiRequest := map[string]interface{}{
			"model": fmt.Sprintf("models/%s", info.UpstreamModelName),
			"content": dto.GeminiChatContent{
				Parts: []dto.GeminiPart{
					{
						Text: input,
					},
				},
			},
		}

		// set specific parameters for different models
		// https://ai.google.dev/api/embeddings?hl=zh-cn#method:-models.embedcontent
		switch info.UpstreamModelName {
		case "text-embedding-004", "gemini-embedding-exp-03-07", "gemini-embedding-001":
			// Only newer models introduced after 2024 support OutputDimensionality
			dimensions := lo.FromPtrOr(request.Dimensions, 0)
			if dimensions > 0 {
				geminiRequest["outputDimensionality"] = dimensions
			}
		}
		geminiRequests = append(geminiRequests, geminiRequest)
	}

	return map[string]interface{}{
		"requests": geminiRequests,
	}, nil
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	result, err := relayconvert.ConvertRequest(c, info, types.RelayFormatGemini, &request)
	if err != nil {
		return nil, err
	}
	geminiRequest, ok := result.Value.(*dto.GeminiChatRequest)
	if !ok {
		return nil, fmt.Errorf("expected Gemini generateContent request, got %T", result.Value)
	}
	return geminiRequest, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	provider := common.GetContextKeyString(c, rootconstant.ContextKeyUpstreamOAuthProvider)
	if provider == "gemini-cli" || provider == "antigravity" {
		raw, err := io.ReadAll(requestBody)
		if err != nil {
			return nil, err
		}
		var request any
		if err = common.Unmarshal(raw, &request); err != nil {
			return nil, err
		}
		metadata, _ := common.GetContextKeyType[*service.UpstreamOAuthTokenPayload](c, rootconstant.ContextKeyUpstreamOAuthMetadata)
		project := ""
		if metadata != nil {
			project = metadata.ProjectID
		}
		wrapped, err := common.Marshal(map[string]any{"model": info.UpstreamModelName, "project": project, "request": request})
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewReader(wrapped)
	}
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	if info.UpstreamOAuthProvider == "gemini-cli" || info.UpstreamOAuthProvider == "antigravity" {
		if unwrapErr := unwrapCodeAssistResponse(resp, info.IsStream); unwrapErr != nil {
			return nil, types.NewError(unwrapErr, types.ErrorCodeBadResponseBody)
		}
	}
	if info.RelayMode == constant.RelayModeResponses {
		if info.IsStream {
			return GeminiResponsesStreamHandler(c, info, resp)
		}
		return GeminiResponsesHandler(c, info, resp)
	}

	if info.RelayMode == constant.RelayModeGemini {
		if strings.Contains(info.RequestURLPath, ":embedContent") ||
			strings.Contains(info.RequestURLPath, ":batchEmbedContents") {
			return NativeGeminiEmbeddingHandler(c, resp, info)
		}
		if info.IsStream {
			return GeminiTextGenerationStreamHandler(c, info, resp)
		} else {
			return GeminiTextGenerationHandler(c, info, resp)
		}
	}

	if strings.HasPrefix(info.UpstreamModelName, "imagen") {
		return GeminiImageHandler(c, info, resp)
	}

	// check if the model is an embedding model
	if strings.HasPrefix(info.UpstreamModelName, "text-embedding") ||
		strings.HasPrefix(info.UpstreamModelName, "embedding") ||
		strings.HasPrefix(info.UpstreamModelName, "gemini-embedding") {
		return GeminiEmbeddingHandler(c, info, resp)
	}

	if info.IsStream {
		return GeminiChatStreamHandler(c, info, resp)
	} else {
		return GeminiChatHandler(c, info, resp)
	}

}

func unwrapCodeAssistResponse(resp *http.Response, stream bool) error {
	if resp == nil || resp.Body == nil {
		return nil
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if stream {
		lines := strings.Split(string(raw), "\n")
		for i, line := range lines {
			prefix := ""
			data := strings.TrimSpace(line)
			if strings.HasPrefix(data, "data:") {
				prefix, data = "data: ", strings.TrimSpace(strings.TrimPrefix(data, "data:"))
			}
			if data == "" || data == "[DONE]" {
				continue
			}
			if unwrapped, ok := unwrapCodeAssistJSON([]byte(data)); ok {
				lines[i] = prefix + string(unwrapped)
			}
		}
		raw = []byte(strings.Join(lines, "\n"))
	} else if unwrapped, ok := unwrapCodeAssistJSON(raw); ok {
		raw = unwrapped
	}
	resp.Body = io.NopCloser(bytes.NewReader(raw))
	resp.ContentLength = int64(len(raw))
	return nil
}

func unwrapCodeAssistJSON(raw []byte) ([]byte, bool) {
	var envelope struct {
		Response any `json:"response"`
	}
	if common.Unmarshal(raw, &envelope) != nil || envelope.Response == nil {
		return raw, false
	}
	unwrapped, err := common.Marshal(envelope.Response)
	return unwrapped, err == nil
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
