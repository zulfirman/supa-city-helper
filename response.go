package zhelper

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Content interface{} `json:"content"`
	Others  interface{} `json:"others"`
	Path    string      `json:"path"`
}
