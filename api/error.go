package api

const (
	Success = 0
	// ParseError begin for json-rpc 2.0
	ParseError      = -32700
	InvalidRequest  = -32600
	MethodNotFound  = -32601
	InvalidParams   = -32602
	InternalError   = -32603
	ReserveMinError = -32099
	ReserveMaxError = -32000
	// ReserveMaxError end for json-rpc 2.0
)

var SysCodeMap = map[int]string{
	ParseError:     "Parse error",
	InvalidRequest: "Invalid request",
	MethodNotFound: "Method not found",
	InvalidParams:  "Invalid params",
	InternalError:  "Internal error",
}
