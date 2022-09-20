package httpapi

const (
	Success         = 0
	ParseError      = -32700
	InvalidRequest  = -32600
	MethodNotFound  = -32601
	InvalidParams   = -32602
	InternalError   = -32603
	ReserveMinError = -32099
	ReserveMaxError = -32000
)

var SysCodeMap = map[int]string{
	ParseError:     "Parse error",
	InvalidRequest: "Invalid request",
	MethodNotFound: "Method not found",
	InvalidParams:  "Invalid params",
	InternalError:  "Internal error",
}

func CheckCode(code int) bool {
	if _, ok := SysCodeMap[code]; ok {
		return false
	}

	if code >= ReserveMinError && code <= ReserveMaxError {
		return false
	}

	return true
}
