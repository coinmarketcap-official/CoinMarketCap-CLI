package api

type CoinInfoURLs struct {
	Website      []string `json:"website,omitempty"`
	TechnicalDoc []string `json:"technical_doc,omitempty"`
	Twitter      []string `json:"twitter,omitempty"`
	Reddit       []string `json:"reddit,omitempty"`
	MessageBoard []string `json:"message_board,omitempty"`
	Announcement []string `json:"announcement,omitempty"`
	Chat         []string `json:"chat,omitempty"`
	Explorer     []string `json:"explorer,omitempty"`
	SourceCode   []string `json:"source_code,omitempty"`
}
