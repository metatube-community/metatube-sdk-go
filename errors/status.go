package errors

import "net/http"

var (
	ErrBadRequest                   = &HTTPError{Code: http.StatusBadRequest}                   // RFC 7231, 6.5.1
	ErrUnauthorized                 = &HTTPError{Code: http.StatusUnauthorized}                 // RFC 7235, 3.1
	ErrPaymentRequired              = &HTTPError{Code: http.StatusPaymentRequired}              // RFC 7231, 6.5.2
	ErrForbidden                    = &HTTPError{Code: http.StatusForbidden}                    // RFC 7231, 6.5.3
	ErrNotFound                     = &HTTPError{Code: http.StatusNotFound}                     // RFC 7231, 6.5.4
	ErrMethodNotAllowed             = &HTTPError{Code: http.StatusMethodNotAllowed}             // RFC 7231, 6.5.5
	ErrNotAcceptable                = &HTTPError{Code: http.StatusNotAcceptable}                // RFC 7231, 6.5.6
	ErrProxyAuthRequired            = &HTTPError{Code: http.StatusProxyAuthRequired}            // RFC 7235, 3.2
	ErrRequestTimeout               = &HTTPError{Code: http.StatusRequestTimeout}               // RFC 7231, 6.5.7
	ErrConflict                     = &HTTPError{Code: http.StatusConflict}                     // RFC 7231, 6.5.8
	ErrGone                         = &HTTPError{Code: http.StatusGone}                         // RFC 7231, 6.5.9
	ErrLengthRequired               = &HTTPError{Code: http.StatusLengthRequired}               // RFC 7231, 6.5.10
	ErrPreconditionFailed           = &HTTPError{Code: http.StatusPreconditionFailed}           // RFC 7232, 4.2
	ErrRequestEntityTooLarge        = &HTTPError{Code: http.StatusRequestEntityTooLarge}        // RFC 7231, 6.5.11
	ErrRequestURITooLong            = &HTTPError{Code: http.StatusRequestURITooLong}            // RFC 7231, 6.5.12
	ErrUnsupportedMediaType         = &HTTPError{Code: http.StatusUnsupportedMediaType}         // RFC 7231, 6.5.13
	ErrRequestedRangeNotSatisfiable = &HTTPError{Code: http.StatusRequestedRangeNotSatisfiable} // RFC 7233, 4.4
	ErrExpectationFailed            = &HTTPError{Code: http.StatusExpectationFailed}            // RFC 7231, 6.5.14
	ErrTeapot                       = &HTTPError{Code: http.StatusTeapot}                       // RFC 7168, 2.3.3
	ErrMisdirectedRequest           = &HTTPError{Code: http.StatusMisdirectedRequest}           // RFC 7540, 9.1.2
	ErrUnprocessableEntity          = &HTTPError{Code: http.StatusUnprocessableEntity}          // RFC 4918, 11.2
	ErrLocked                       = &HTTPError{Code: http.StatusLocked}                       // RFC 4918, 11.3
	ErrFailedDependency             = &HTTPError{Code: http.StatusFailedDependency}             // RFC 4918, 11.4
	ErrTooEarly                     = &HTTPError{Code: http.StatusTooEarly}                     // RFC 8470, 5.2.
	ErrUpgradeRequired              = &HTTPError{Code: http.StatusUpgradeRequired}              // RFC 7231, 6.5.15
	ErrPreconditionRequired         = &HTTPError{Code: http.StatusPreconditionRequired}         // RFC 6585, 3
	ErrTooManyRequests              = &HTTPError{Code: http.StatusTooManyRequests}              // RFC 6585, 4
	ErrRequestHeaderFieldsTooLarge  = &HTTPError{Code: http.StatusRequestHeaderFieldsTooLarge}  // RFC 6585, 5
	ErrUnavailableForLegalReasons   = &HTTPError{Code: http.StatusUnavailableForLegalReasons}   // RFC 7725, 3

	ErrInternalServerError           = &HTTPError{Code: http.StatusInternalServerError}           // RFC 7231, 6.6.1
	ErrNotImplemented                = &HTTPError{Code: http.StatusNotImplemented}                // RFC 7231, 6.6.2
	ErrBadGateway                    = &HTTPError{Code: http.StatusBadGateway}                    // RFC 7231, 6.6.3
	ErrServiceUnavailable            = &HTTPError{Code: http.StatusServiceUnavailable}            // RFC 7231, 6.6.4
	ErrGatewayTimeout                = &HTTPError{Code: http.StatusGatewayTimeout}                // RFC 7231, 6.6.5
	ErrHTTPVersionNotSupported       = &HTTPError{Code: http.StatusHTTPVersionNotSupported}       // RFC 7231, 6.6.6
	ErrVariantAlsoNegotiates         = &HTTPError{Code: http.StatusVariantAlsoNegotiates}         // RFC 2295, 8.1
	ErrInsufficientStorage           = &HTTPError{Code: http.StatusInsufficientStorage}           // RFC 4918, 11.5
	ErrLoopDetected                  = &HTTPError{Code: http.StatusLoopDetected}                  // RFC 5842, 7.2
	ErrNotExtended                   = &HTTPError{Code: http.StatusNotExtended}                   // RFC 2774, 7
	ErrNetworkAuthenticationRequired = &HTTPError{Code: http.StatusNetworkAuthenticationRequired} // RFC 6585, 6
)
