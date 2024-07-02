package response

import (
	"certification/constant"

	"github.com/google/uuid"
)

type MessageDataResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type LoginSuccessResponse struct {
	ID    uuid.UUID `json:"id"`
	Token string    `json:"token"`
	Email string    `json:"email"`
}

func LoginFailResponseBody() MessageResponse {
	return MessageResponse{
		Status:  constant.ERROR,
		Message: "Invalid email or password",
	}
}

func InvalidPasswordResponseBody() MessageResponse {
	return MessageResponse{
		Status:  constant.ERROR,
		Message: "Invalid password",
	}
}

func LoginSuccessResponseBody(
	id uuid.UUID,
	token string,
	email string,
) MessageDataResponse {
	return MessageDataResponse{
		Status:  constant.SUCCESS,
		Message: constant.SuccessLogIn,
		Data: LoginSuccessResponse{
			ID:    id,
			Token: token,
			Email: email,
		},
	}
}

func LogoutFailResponseBody(message string) MessageResponse {
	return MessageResponse{
		Status:  constant.ErrorLogOut,
		Message: message,
	}
}

func LogoutSuccessResponseBody() MessageResponse {
	return MessageResponse{
		Status:  constant.SUCCESS,
		Message: constant.SuccessLogOut,
	}
}
