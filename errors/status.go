package errors

import (
	"net/http"
)

var statusCode = map[string]int{
	"Continue":            http.StatusContinue,
	"Switching Protocols": http.StatusSwitchingProtocols,
	"Processing":          http.StatusProcessing,
	"Early Hints":         http.StatusEarlyHints,

	"OK":                            http.StatusOK,
	"Created":                       http.StatusCreated,
	"Accepted":                      http.StatusAccepted,
	"Non-Authoritative Information": http.StatusNonAuthoritativeInfo,
	"No Content":                    http.StatusNoContent,
	"Reset Content":                 http.StatusResetContent,
	"Partial Content":               http.StatusPartialContent,
	"Multi-Status":                  http.StatusMultiStatus,
	"Already Reported":              http.StatusAlreadyReported,
	"IM Used":                       http.StatusIMUsed,

	"Multiple Choices":   http.StatusMultipleChoices,
	"Moved Permanently":  http.StatusMovedPermanently,
	"Found":              http.StatusFound,
	"See Other":          http.StatusSeeOther,
	"Not Modified":       http.StatusNotModified,
	"Use Proxy":          http.StatusUseProxy,
	"Temporary Redirect": http.StatusTemporaryRedirect,
	"Permanent Redirect": http.StatusPermanentRedirect,

	"Bad Request":                     http.StatusBadRequest,
	"Unauthorized":                    http.StatusUnauthorized,
	"Payment Required":                http.StatusPaymentRequired,
	"Forbidden":                       http.StatusForbidden,
	"Not Found":                       http.StatusNotFound,
	"Method Not Allowed":              http.StatusMethodNotAllowed,
	"Not Acceptable":                  http.StatusNotAcceptable,
	"Proxy Authentication Required":   http.StatusProxyAuthRequired,
	"Request Timeout":                 http.StatusRequestTimeout,
	"Conflict":                        http.StatusConflict,
	"Gone":                            http.StatusGone,
	"Length Required":                 http.StatusLengthRequired,
	"Precondition Failed":             http.StatusPreconditionFailed,
	"Request Entity Too Large":        http.StatusRequestEntityTooLarge,
	"Request URI Too Long":            http.StatusRequestURITooLong,
	"Unsupported Media Type":          http.StatusUnsupportedMediaType,
	"Requested Range Not Satisfiable": http.StatusRequestedRangeNotSatisfiable,
	"Expectation Failed":              http.StatusExpectationFailed,
	"I'm a teapot":                    http.StatusTeapot,
	"Misdirected Request":             http.StatusMisdirectedRequest,
	"Unprocessable Entity":            http.StatusUnprocessableEntity,
	"Locked":                          http.StatusLocked,
	"Failed Dependency":               http.StatusFailedDependency,
	"Too Early":                       http.StatusTooEarly,
	"Upgrade Required":                http.StatusUpgradeRequired,
	"Precondition Required":           http.StatusPreconditionRequired,
	"Too Many Requests":               http.StatusTooManyRequests,
	"Request Header Fields Too Large": http.StatusRequestHeaderFieldsTooLarge,
	"Unavailable For Legal Reasons":   http.StatusUnavailableForLegalReasons,

	"Internal Server Error":           http.StatusInternalServerError,
	"Not Implemented":                 http.StatusNotImplemented,
	"Bad Gateway":                     http.StatusBadGateway,
	"Service Unavailable":             http.StatusServiceUnavailable,
	"Gateway Timeout":                 http.StatusGatewayTimeout,
	"HTTP Version Not Supported":      http.StatusHTTPVersionNotSupported,
	"Variant Also Negotiates":         http.StatusVariantAlsoNegotiates,
	"Insufficient Storage":            http.StatusInsufficientStorage,
	"Loop Detected":                   http.StatusLoopDetected,
	"Not Extended":                    http.StatusNotExtended,
	"Network Authentication Required": http.StatusNetworkAuthenticationRequired,
}

// StatusCode is a reverse function of http.StatusText.
func StatusCode(text any) int {
	switch v := text.(type) {
	case string:
		return statusCode[v]
	case error:
		return statusCode[v.Error()]
	default:
		return 0
	}
}
