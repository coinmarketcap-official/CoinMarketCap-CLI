package api

type NewsLatestResponse struct {
	Data []NewsArticle `json:"data"`
}

type NewsArticle struct {
	Title      string         `json:"title"`
	Subtitle   string         `json:"subtitle"`
	SourceName string         `json:"source_name"`
	SourceURL  string         `json:"source_url"`
	ReleasedAt string         `json:"released_at"`
	CreatedAt  string         `json:"created_at"`
	Assets     []NewsAssetRef `json:"assets"`
	Type       string         `json:"type"`
	NewsType   string         `json:"news_type"`
	Cover      string         `json:"cover"`
	Language   string         `json:"language"`
}

type NewsAssetRef struct {
	Symbol string `json:"symbol"`
}
