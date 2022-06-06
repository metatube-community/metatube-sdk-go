package errors

import "net/http"

var (
	ErrBadRequest                   = FromCode(http.StatusBadRequest)                   // RFC 7231, 6.5.1
	ErrUnauthorized                 = FromCode(http.StatusUnauthorized)                 // RFC 7235, 3.1
	ErrPaymentRequired              = FromCode(http.StatusPaymentRequired)              // RFC 7231, 6.5.2
	ErrForbidden                    = FromCode(http.StatusForbidden)                    // RFC 7231, 6.5.3
	ErrNotFound                     = FromCode(http.StatusNotFound)                     // RFC 7231, 6.5.4
	ErrMethodNotAllowed             = FromCode(http.StatusMethodNotAllowed)             // RFC 7231, 6.5.5
	ErrNotAcceptable                = FromCode(http.StatusNotAcceptable)                // RFC 7231, 6.5.6
	ErrProxyAuthRequired            = FromCode(http.StatusProxyAuthRequired)            // RFC 7235, 3.2
	ErrRequestTimeout               = FromCode(http.StatusRequestTimeout)               // RFC 7231, 6.5.7
	ErrConflict                     = FromCode(http.StatusConflict)                     // RFC 7231, 6.5.8
	ErrGone                         = FromCode(http.StatusGone)                         // RFC 7231, 6.5.9
	ErrLengthRequired               = FromCode(http.StatusLengthRequired)               // RFC 7231, 6.5.10
	ErrPreconditionFailed           = FromCode(http.StatusPreconditionFailed)           // RFC 7232, 4.2
	ErrRequestEntityTooLarge        = FromCode(http.StatusRequestEntityTooLarge)        // RFC 7231, 6.5.11
	ErrRequestURITooLong            = FromCode(http.StatusRequestURITooLong)            // RFC 7231, 6.5.12
	ErrUnsupportedMediaType         = FromCode(http.StatusUnsupportedMediaType)         // RFC 7231, 6.5.13
	ErrRequestedRangeNotSatisfiable = FromCode(http.StatusRequestedRangeNotSatisfiable) // RFC 7233, 4.4
	ErrExpectationFailed            = FromCode(http.StatusExpectationFailed)            // RFC 7231, 6.5.14
	ErrTeapot                       = FromCode(http.StatusTeapot)                       // RFC 7168, 2.3.3
	ErrMisdirectedRequest           = FromCode(http.StatusMisdirectedRequest)           // RFC 7540, 9.1.2
	ErrUnprocessableEntity          = FromCode(http.StatusUnprocessableEntity)          // RFC 4918, 11.2
	ErrLocked                       = FromCode(http.StatusLocked)                       // RFC 4918, 11.3
	ErrFailedDependency             = FromCode(http.StatusFailedDependency)             // RFC 4918, 11.4
	ErrTooEarly                     = FromCode(http.StatusTooEarly)                     // RFC 8470, 5.2.
	ErrUpgradeRequired              = FromCode(http.StatusUpgradeRequired)              // RFC 7231, 6.5.15
	ErrPreconditionRequired         = FromCode(http.StatusPreconditionRequired)         // RFC 6585, 3
	ErrTooManyRequests              = FromCode(http.StatusTooManyRequests)              // RFC 6585, 4
	ErrRequestHeaderFieldsTooLarge  = FromCode(http.StatusRequestHeaderFieldsTooLarge)  // RFC 6585, 5
	ErrUnavailableForLegalReasons   = FromCode(http.StatusUnavailableForLegalReasons)   // RFC 7725, 3

	ErrInternalServerError           = FromCode(http.StatusInternalServerError)           // RFC 7231, 6.6.1
	ErrNotImplemented                = FromCode(http.StatusNotImplemented)                // RFC 7231, 6.6.2
	ErrBadGateway                    = FromCode(http.StatusBadGateway)                    // RFC 7231, 6.6.3
	ErrServiceUnavailable            = FromCode(http.StatusServiceUnavailable)            // RFC 7231, 6.6.4
	ErrGatewayTimeout                = FromCode(http.StatusGatewayTimeout)                // RFC 7231, 6.6.5
	ErrHTTPVersionNotSupported       = FromCode(http.StatusHTTPVersionNotSupported)       // RFC 7231, 6.6.6
	ErrVariantAlsoNegotiates         = FromCode(http.StatusVariantAlsoNegotiates)         // RFC 2295, 8.1
	ErrInsufficientStorage           = FromCode(http.StatusInsufficientStorage)           // RFC 4918, 11.5
	ErrLoopDetected                  = FromCode(http.StatusLoopDetected)                  // RFC 5842, 7.2
	ErrNotExtended                   = FromCode(http.StatusNotExtended)                   // RFC 2774, 7
	ErrNetworkAuthenticationRequired = FromCode(http.StatusNetworkAuthenticationRequired) // RFC 6585, 6
)
