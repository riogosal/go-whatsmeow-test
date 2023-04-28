package chatgpt

type Choice struct {
	Message      Message      `json:"text"`
	Index        int          `json:"index"`
	FinishReason FinishReason `json:"finish_reason"`
}
