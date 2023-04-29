package chatgpt

type ChatGPTClient interface {
	ChatCompletion(prompt string) (string, error)

	WithSystemPrompt(system_prompt string)
}
