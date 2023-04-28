package chatgpt

// stop: API returned complete model output
// length: Incomplete model output due to max_tokens parameter or token limit
// content_filter: Omitted content due to a flag from our content filters
// null: API response still in progress or incomplete

type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonLength        FinishReason = "length"
	FinishReasonContentFilter FinishReason = "content_filter"
	FinishReasonNull          FinishReason = "null"
)
