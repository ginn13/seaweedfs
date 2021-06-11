package s3err

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/chrislusf/seaweedfs/weed/glog"
	"net/http"
	"strconv"
	"time"
)

type mimeType string

const (
	mimeNone mimeType = ""
	MimeXML  mimeType = "application/xml"
)

func WriteXMLResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	WriteResponse(w, statusCode, EncodeXMLResponse(response), MimeXML)
}

func WriteEmptyResponse(w http.ResponseWriter, statusCode int) {
	WriteResponse(w, statusCode, []byte{}, mimeNone)
}

func WriteErrorResponse(w http.ResponseWriter, errorCode ErrorCode, r *http.Request) {
	apiError := GetAPIError(errorCode)
	errorResponse := getRESTErrorResponse(apiError, r.URL.Path)
	encodedErrorResponse := EncodeXMLResponse(errorResponse)
	WriteResponse(w, apiError.HTTPStatusCode, encodedErrorResponse, MimeXML)
}

func getRESTErrorResponse(err APIError, resource string) RESTErrorResponse {
	return RESTErrorResponse{
		Code:      err.Code,
		Message:   err.Description,
		Resource:  resource,
		RequestID: fmt.Sprintf("%d", time.Now().UnixNano()),
	}
}

// Encodes the response headers into XML format.
func EncodeXMLResponse(response interface{}) []byte {
	var bytesBuffer bytes.Buffer
	bytesBuffer.WriteString(xml.Header)
	e := xml.NewEncoder(&bytesBuffer)
	e.Encode(response)
	return bytesBuffer.Bytes()
}

func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("x-amz-request-id", fmt.Sprintf("%d", time.Now().UnixNano()))
	w.Header().Set("Accept-Ranges", "bytes")
}

func WriteResponse(w http.ResponseWriter, statusCode int, response []byte, mType mimeType) {
	setCommonHeaders(w)
	if response != nil {
		w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	}
	if mType != mimeNone {
		w.Header().Set("Content-Type", string(mType))
	}
	w.WriteHeader(statusCode)
	if response != nil {
		glog.V(4).Infof("status %d %s: %s", statusCode, mType, string(response))
		_, err := w.Write(response)
		if err != nil {
			glog.V(0).Infof("write err: %v", err)
		}
		w.(http.Flusher).Flush()
	}
}

// If none of the http routes match respond with MethodNotAllowed
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(0).Infof("unsupported %s %s", r.Method, r.RequestURI)
	WriteErrorResponse(w, ErrMethodNotAllowed, r)
}
