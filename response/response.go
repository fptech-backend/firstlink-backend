package response

import (
	"certification/constant"
	"fmt"
)

type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type DataResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message" default:""`
	Data    interface{} `json:"data"`
}

type ResponseCreated struct {
	ID string `json:"id"`
}

type Response struct {
	Status  string       `json:"status"`
	Message string       `json:"message" default:""`
	Data    *interface{} `json:"data"`
}

func SuccessResponseBody(message string) MessageResponse {
	return MessageResponse{
		Status:  constant.SUCCESS,
		Message: message,
	}
}

func ErrorResponseBody(errorMsg string, args ...interface{}) MessageResponse {
	errorMsg = GetMessage(errorMsg, args)

	return MessageResponse{
		Status:  constant.ERROR,
		Message: errorMsg,
	}
}

func DataResponseBody(data interface{}, message string) DataResponse {
	return DataResponse{
		Status:  constant.SUCCESS,
		Message: message,
		Data:    data,
	}
}

func AccessDeniedResponseBody(id string) MessageResponse {
	return MessageResponse{
		Status:  constant.ACCESS_DENIED,
		Message: "User with ID " + id + " does not have sufficient permission to access this module ",
	}
}

// GetMessage format with Sprint, Sprintf, or neither.
func GetMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}

	return fmt.Sprint(fmtArgs...)
}
