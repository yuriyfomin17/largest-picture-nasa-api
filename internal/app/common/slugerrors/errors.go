package slugerrors

type ErrorType struct {
	t string
}

var (
	ErrorTypeUnknown    = ErrorType{"unknown"}
	ErrorTypeBadRequest = ErrorType{"bad-request"}
	ErrorTypeNotFound   = ErrorType{"not-found"}
)

type SlugError struct {
	error     string
	slug      string
	errorType ErrorType
}

func (s SlugError) Error() string {
	return s.error
}

func (s SlugError) Slug() string {
	return s.slug
}

func (s SlugError) ErrorType() ErrorType {
	return s.errorType
}

func NewUnknownError(errorMessage, slug string) SlugError {
	return SlugError{
		error:     errorMessage,
		slug:      slug,
		errorType: ErrorTypeUnknown,
	}
}

// NewBadRequestError creates a SlugError of type "bad-request".
func NewBadRequestError(errorMessage, slug string) SlugError {
	return SlugError{
		error:     errorMessage,
		slug:      slug,
		errorType: ErrorTypeBadRequest,
	}
}

// NewNotFoundError creates a SlugError of type "not-found".
func NewNotFoundError(errorMessage, slug string) SlugError {
	return SlugError{
		error:     errorMessage,
		slug:      slug,
		errorType: ErrorTypeNotFound,
	}
}
