package encode

type ValidationError struct {
	Message string
}

func (v ValidationError) Error() string {
	return v.Message
}
