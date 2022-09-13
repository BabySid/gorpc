package http

const (
	Success         = 0
	ParseError      = -32700
	InvalidRequest  = -32600
	MethodNotFound  = -32601
	InvalidParams   = -32602
	InternalError   = -32603
	ReserveMinError = -32000
	ReserveMaxError = -32099
)

var sysCodeMap = map[int]string{
	ParseError:     "Parse error",
	InvalidRequest: "Invalid request",
	MethodNotFound: "Method not found",
	InvalidParams:  "Invalid params",
	InternalError:  "Internal error",
}

func checkCode(code int) bool {
	if _, ok := sysCodeMap[code]; ok {
		return false
	}

	if code >= ReserveMinError && code <= ReserveMaxError {
		return false
	}

	return true
}
