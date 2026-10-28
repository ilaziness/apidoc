package httpapi

// Spec is the persisted HTTP interface document (JSON in store.Node.HTTP).
type Spec struct {
	Method             string `json:"method"`
	URL                string `json:"url"`
	PathParams         []KV   `json:"pathParams,omitempty"`
	Query              []KV   `json:"query,omitempty"`
	Headers            []KV   `json:"headers,omitempty"`
	Body               Body   `json:"body"`
	Auth               *Auth  `json:"auth,omitempty"` // nil = inherit collection default
	TimeoutSec         int    `json:"timeoutSec"`
	FollowRedirects    bool   `json:"followRedirects"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
	Description        string `json:"description,omitempty"`
}

// KV is a key/value row with an enable flag.
type KV struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// Body kinds: none | json | form | multipart | raw | binary
type Body struct {
	Kind        string `json:"kind"`
	JSON        string `json:"json,omitempty"`
	Form        []KV   `json:"form,omitempty"`
	Multipart   []KV   `json:"multipart,omitempty"` // value may be file:path for files
	Raw         string `json:"raw,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	BinaryPath  string `json:"binaryPath,omitempty"`
}

// Auth kinds: none | bearer | basic | apikey
type Auth struct {
	Kind  string `json:"kind"`
	Token string `json:"token,omitempty"`
	User  string `json:"user,omitempty"`
	Pass  string `json:"pass,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
	In    string `json:"in,omitempty"` // header | query
}

// DefaultSpec returns PRD transport defaults for a new HTTP node.
func DefaultSpec() Spec {
	return Spec{
		Method:             "GET",
		URL:                "",
		Body:               Body{Kind: "none"},
		TimeoutSec:         60,
		FollowRedirects:    true,
		InsecureSkipVerify: false,
	}
}
