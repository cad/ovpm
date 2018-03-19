package errors

// Error ...
type Error struct {
	Message string                 `json:"message"`
	Args    map[string]interface{} `json:"args"`
	Code    int                    `json:"code"`
}

func (e Error) Error() string {
	return e.Message
}
