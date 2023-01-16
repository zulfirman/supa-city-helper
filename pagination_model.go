package zhelper

type Pagination struct {
	Limit  int    `json:"limit"`
	Page   int    `json:"page"`
	Sort   string `json:"sort"`
	Search string `json:"search"`
	Field  string `json:"field"`
}
