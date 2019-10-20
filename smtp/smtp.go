package smtp

// Error holds error messages from Smtp commands with a message and code
type Error struct {
	msg  string
	code int
}

// NewError creates a new Error
func NewError(code int, msg string) *Error {
	result := Error{}
	result.msg = msg
	result.code = code
	return &result
}
