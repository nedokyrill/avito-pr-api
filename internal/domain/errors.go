package domain

import "github.com/nedokyrill/avito-pr-api/internal/generated"

// Реэкспорт типов из generated для использования в доменной логике

type ErrorResponseErrorCode = generated.ErrorResponseErrorCode

type ErrorResponse = generated.ErrorResponse

// Реэкспорт констант кодов ошибок из OpenAPI
const (
	NoCandidate ErrorResponseErrorCode = generated.NOCANDIDATE
	NotAssigned ErrorResponseErrorCode = generated.NOTASSIGNED
	NotFound    ErrorResponseErrorCode = generated.NOTFOUND
	PrExists    ErrorResponseErrorCode = generated.PREXISTS
	PrMerged    ErrorResponseErrorCode = generated.PRMERGED
	TeamExists  ErrorResponseErrorCode = generated.TEAMEXISTS
)

// Кастомные 400 и 500
const (
	InvalidRequest ErrorResponseErrorCode = "INVALID_REQUEST"
	InternalError  ErrorResponseErrorCode = "INTERNAL_ERROR"
)

// Ошибки между репо и сервис слоями
const (
	TeamNotExistsErr string = "team does not exist"
	NoUsersInTeamErr string = "no users in team"
)

func NewErrorResponse(code ErrorResponseErrorCode, message string) ErrorResponse {
	return ErrorResponse{
		Error: struct {
			Code    ErrorResponseErrorCode `json:"code"`
			Message string                 `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	}
}
