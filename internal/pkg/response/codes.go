package response

const (
	CodeValidation         = "VALIDATION_ERROR"
	CodeDuplicateEmail     = "DUPLICATE_EMAIL"
	CodeInvalidCredential  = "INVALID_CREDENTIAL"
	CodeNotFound           = "NOT_FOUND"
	CodeInternalServer     = "INTERNAL_SERVER_ERROR"
	CodeIdempotencyFailed  = "IDEMPOTENCY_FAILED"
	CodeIdempotencyExists  = "IDEMPOTENCY_KEY_EXISTS"
	CodeForbidden          = "FORBIDDEN"
)
