package constant

const (
	// Error message
	ErrorDeleteRecord   = "Failed to delete the record "
	ErrorDuplicateEntry = "data already exists "
	ErrorEmptyValue     = "Empty value is not allowed"
	ErrorInvalidValue   = "invalid value "
	ErrorInvalidID      = "invalid ID "
	ErrorLogOut         = "Failed to log out"
	ErrorWriteAccess    = "WRITE access cannot be granted while READ access is set to false"

	// Success message
	SuccessCreateRecord = "Successfully created"
	SuccessDeleteRecord = "Successfully deleted"
	SuccessUpdateRecord = "Successfully updated"
	SuccessLogIn        = "Successfully logged in"
	SuccessLogOut       = "Successfully logged out"
	SuccessSignUp       = "Successfully signed up"
	SuccessValidate     = "Successfully validated"
)
