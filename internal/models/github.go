package models

// GitHub Push Webhook ペイロード
type GitHubPushPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
	Sender struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"sender"`
	Commits []struct {
		ID        string   `json:"id"`
		Message   string   `json:"message"`
		Author    struct {
			Name string `json:"name"`
		} `json:"author"`
		URL       string   `json:"url"`
		Added     []string `json:"added"`
		Removed   []string `json:"removed"`
		Modified  []string `json:"modified"`
	} `json:"commits"`
}

// GitHub Pull Request Webhook ペイロード
type GitHubPRPayload struct {
	Action      string `json:"action"` // opened, closed 等
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		Merged  bool   `json:"merged"`
		User    struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
		MergedBy struct {
			Login string `json:"login"`
		} `json:"merged_by"`
	} `json:"pull_request"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
}