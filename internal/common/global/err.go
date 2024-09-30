package global

const (
	ErrTokenExpired     = "token is expired"
	ErrTokenRevoked     = "token has been revoked"
	ErrTokenNotValidYet = "token not active yet"
	ErrTokenMalformed   = "that's not even a token"
	ErrTokenInvalid     = "couldn't handle this token"
	ErrEmptyAToken      = "token is empty"
	ErrServerNotOK      = "服务内部错误"

	// ErrMissingSecretKey indicates Secret key is required
	ErrMissingSecretKey = "secret key is required"

	// ErrForbidden when HTTP status 403 is given
	ErrForbidden = "you don't have permission to access this resource"

	// ErrFailedAuthentication indicates authentication failed, could be faulty username or password
	ErrFailedAuthentication = "用户名或密码错误"

	// ErrFailedTokenCreation indicates JWT Token failed to create, reason unknown
	ErrFailedTokenCreation = "failed to create JWT Token"

	// ErrMissingExpField missing exp field in token
	ErrMissingExpField = "missing exp field"

	// ErrWrongFormatOfExp field must be float64 format
	ErrWrongFormatOfExp = "exp must be float64 format"

	ErrInvalidVerificationode = "验证码错误"

	ErrEmptyOptsForGromDelete = "delete operation requires at least one condition to avoid accidental data loss"
	ErrEmptyOptsForGromUptdae = "update operation requires at least one condition to avoid accidental data loss"
)
