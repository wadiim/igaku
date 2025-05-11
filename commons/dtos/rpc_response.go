package dtos

import (
	"encoding/json"
)

type RPCResponse struct {
	Data	json.RawMessage	`json:data`
	Error	*RPCError	`json:error`
}

type RPCError struct {
	Code	string		`json:code`
	Message	string		`json:message`
}
