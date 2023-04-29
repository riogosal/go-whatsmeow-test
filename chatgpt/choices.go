package chatgpt

type Choice struct {
	Message      Message      `json:"message"`
	Index        int          `json:"index"`
	FinishReason FinishReason `json:"finish_reason"`
}
