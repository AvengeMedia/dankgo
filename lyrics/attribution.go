package lyrics

// Attribution credits the source of a Result. Text is the wording the source asks for, if any.
type Attribution struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Text string `json:"text,omitempty"`
}
