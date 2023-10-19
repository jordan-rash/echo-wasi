package main

import (
	echo "echo/gen"
	"encoding/json"
	"net/http"
	"strings"
)

type EchoResponse struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	QueryString string `json:"queryString,omitempty"`
	Body        string `json:"body,omitempty"`
}

type Echo struct{}

func (g *Echo) Handle(req echo.WasiHttpIncomingHandlerIncomingRequest, resp echo.WasiHttpTypesResponseOutparam) {
	er := new(EchoResponse)

	method := echo.WasiHttpTypesIncomingRequestMethod(req)
	switch method {
	case echo.WasiHttpTypesMethodGet():
		er.Method = "GET"
	case echo.WasiHttpTypesMethodPost():
		er.Method = "POST"
	case echo.WasiHttpTypesMethodPut():
		er.Method = "PUT"
	case echo.WasiHttpTypesMethodDelete():
		er.Method = "DELETE"
	case echo.WasiHttpTypesMethodPatch():
		er.Method = "PATCH"
	case echo.WasiHttpTypesMethodConnect():
		er.Method = "CONNECT"
	default:
		er.Method = "OTHER"
	}

	pathWithQuery := echo.WasiHttpTypesIncomingRequestPathWithQuery(req)
	if pathWithQuery.IsNone() {
		return
	}

	splitPathQuery := strings.Split(pathWithQuery.Unwrap(), "?")
	er.Path = splitPathQuery[0]
	if len(splitPathQuery) > 1 {
		er.QueryString = splitPathQuery[1]
	}

	bodyStream := echo.WasiHttpTypesIncomingRequestConsume(req)
	if bodyStream.IsErr() {
		writeHttpResponse(resp, http.StatusInternalServerError, []echo.WasiHttpTypesTuple2StringListU8TT{{F0: "Content-Type", F1: []byte("application/json")}}, []byte("{\"error\":\"failed to read request body\"}"))
		return
	}

	readStream := echo.WasiIoStreamsBlockingRead(bodyStream.Val, 18446744073709551614)
	er.Body = string(readStream.Val.F0)

	echo.WasiLoggingLoggingLog(echo.WasiLoggingLoggingLevelDebug(), "method", er.Method)
	echo.WasiLoggingLoggingLog(echo.WasiLoggingLoggingLevelDebug(), "path", er.Path)
	echo.WasiLoggingLoggingLog(echo.WasiLoggingLoggingLevelDebug(), "queryString", er.QueryString)
	echo.WasiLoggingLoggingLog(echo.WasiLoggingLoggingLevelDebug(), "body", er.Body)

	bBody, err := json.Marshal(er)
	if err != nil {
		writeHttpResponse(resp, http.StatusInternalServerError, []echo.WasiHttpTypesTuple2StringListU8TT{{F0: "Content-Type", F1: []byte("application/json")}}, []byte("{\"error\":\"failed to marshal response\"}"))
		return
	}

	writeHttpResponse(resp, http.StatusOK, []echo.WasiHttpTypesTuple2StringListU8TT{{F0: "Content-Type", F1: []byte("application/json")}}, bBody)
}

func writeHttpResponse(responseOutparam echo.WasiHttpTypesResponseOutparam, statusCode uint16, inHeaders []echo.WasiHttpTypesTuple2StringListU8TT, body []byte) {
	headers := echo.WasiHttpTypesNewFields(inHeaders)

	outgoingResponse := echo.WasiHttpTypesNewOutgoingResponse(statusCode, headers)
	if outgoingResponse.IsErr() {
		return
	}

	outgoingStream := echo.WasiHttpTypesOutgoingResponseWrite(outgoingResponse.Unwrap())
	if outgoingStream.IsErr() {
		return
	}

	cw := echo.WasiIoStreamsCheckWrite(outgoingStream.Val)
	if cw.IsErr() {
		return
	}

	wf := echo.WasiIoStreamsBlockingWriteAndFlush(outgoingStream.Val, body)
	if wf.IsErr() {
		return
	}

	echo.WasiHttpTypesFinishOutgoingStream(outgoingStream.Val)

	outparm := echo.WasiHttpTypesSetResponseOutparam(responseOutparam, outgoingResponse)
	if outparm.IsErr() {
		return
	}
}

func init() {
	mg := new(Echo)
	echo.SetExportsWasiHttpIncomingHandler(mg)
}

//go:generate wit-bindgen tiny-go ./wit -w echo --out-dir=gen
func main() {}
