package dto

const (
	DefaultSearchLimit = 20
	MaxSearchLimit     = 100
)

type SearchProfilesQuery struct {
	Username  string `json:"username,omitempty"`
	FullName  string `json:"full_name,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Age       *int   `json:"age,omitempty"`
	City      string `json:"city,omitempty"`
	Country   string `json:"country,omitempty"`

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

func (sq *SearchProfilesQuery) NormalizePagination() {
	if sq.Limit <= 0 {
		sq.Limit = DefaultSearchLimit
	} else if sq.Limit > MaxSearchLimit {
		sq.Limit = MaxSearchLimit
	}

	if sq.Offset < 0 {
		sq.Offset = 0
	}
}
