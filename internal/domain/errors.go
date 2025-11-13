package domain

import "github.com/nedokyrill/avito-pr-api/internal/generated"

// Реэкспорт типов из generated для использования в доменной логике

type ErrorResponseErrorCode = generated.ErrorResponseErrorCode

type ErrorResponse = generated.ErrorResponse

// Реэкспорт констант кодов ошибок
const (
	NoCandidate ErrorResponseErrorCode = generated.NOCANDIDATE
	NotAssigned ErrorResponseErrorCode = generated.NOTASSIGNED
	NotFound    ErrorResponseErrorCode = generated.NOTFOUND
	PrExists    ErrorResponseErrorCode = generated.PREXISTS
	PrMerged    ErrorResponseErrorCode = generated.PRMERGED
	TeamExists  ErrorResponseErrorCode = generated.TEAMEXISTS
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
