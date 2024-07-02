package constant

const (
	// Redis Status
	CREATED = "created"
	UPDATED = "updated"

	// MongoDB
	MONGO_SET  = "$set"
	MONGO_PUSH = "$push"

	MONGO_ID     = "_id"
	MONGO_FIELDS = "fields"

	// Validator
	VALIDATE       = "validate"
	PATCH_VALIDATE = "patch_validate"

	// Modules
	ACCOUNT = 1
	WALLET  = 2
	PROFILE = 3

	// Access Permission
	READ   = "read"
	WRITE  = "write"  // READ & WRITE
	DELETE = "delete" // READ & WRITE & DELETE

	// Message Status
	SUCCESS       = "success"
	ERROR         = "error"
	ACCESS_DENIED = "Access Denied"

	// Date Time Format
	DATE_FORMAT    = "2006-01-02"
	TIME_FORMAT_24 = "15:04"
	TIME_FORMAT_12 = "3:04 PM"

	// Project Admin's Role
	PROJECT_OWNER = "owner"
	PROJECT_ADMIN = "admin"

	// API KEYS
	DEFAULT_SIZE_PER_API_GROUP = 6
	DEFAULT_SIZE_OF_API_GROUP  = 4
)

// Token Type
const (
	VALIDATION_TOKEN     = "validation"
	RESET_PASSWORD_TOKEN = "reset_password"
	OTP_TOKEN            = "otp"
)

// User Status
type Status string

const (
	ACTIVE    Status = "active"
	INACTIVE  Status = "inactive"
	PENDING   Status = "pending"
	USED      Status = "used"
	COMPLETED Status = "completed"
	OFFCHAIN  Status = "offchain"
	ONCHAIN   Status = "onchain"
	SUBMITTED Status = "submitted"
	DELETED   Status = "deleted"
	FAILED    Status = "failed"
)

// Environment
const (
	ENV_STAGING = "staging"
	ENV_PROD    = "production"
	ENV_LOCAL   = "local"
)

// Field Type with defined string
type FieldType string

const (
	TEXT     FieldType = "text"
	NUMBER   FieldType = "number"
	BOOLEAN  FieldType = "boolean"
	DATE     FieldType = "date"
	DATETIME FieldType = "datetime"
)

// Account Role
type AccountRoleType string

const (
	ROLE_COMPANY AccountRoleType = "company"
	ROLE_USER    AccountRoleType = "user"
)
