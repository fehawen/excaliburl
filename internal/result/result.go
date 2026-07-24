package result

type Result struct {
	File       string `json:"file,omitempty"`
	Processor  string `json:"processor,omitempty"`
	Raw        string `json:"raw"`
	Normalized string `json:"normalized,omitempty"`
}
