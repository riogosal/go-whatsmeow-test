package chatgpt

type ChatGPTClient interface {
	ChatCompletion(prompt string) (string, error)
}
