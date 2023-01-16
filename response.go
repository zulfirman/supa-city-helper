package zhelper

type Response struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	MessageJp string      `json:"messageJp"`
	Content   interface{} `json:"content"`
	Others    interface{} `json:"others"`
	Path      string      `json:"path"`
}
