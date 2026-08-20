package cerrors

import (
	"fmt"
	"io"
	"strconv"

	"transcode-demo/pkg/json"
)

const (
	OK                  Code = 200
	Created             Code = 201
	Accepted            Code = 202
	NoContent           Code = 204
	BadRequest          Code = 400
	Unauthorized        Code = 401
	Forbidden           Code = 403
	NotFound            Code = 404
	MethodNotAllow      Code = 405
	NotAcceptable       Code = 406
	RequestTimeout      Code = 408
	Conflict            Code = 409
	PreconditionFailed  Code = 412
	InternalServerError Code = 500
	ServiceUnavailable  Code = 503
)

var customCodeMessages = map[Code]string{}

func InitCustomCodeMessage(msgs map[Code]string) {
	customCodeMessages = msgs
}

var httpStatusMessages = map[string]string{
	"100": "Continue",
	"101": "Switching Protocols",
	"102": "Processing (WebDAV)",
	"103": "Early Hints Experimental",
	"200": "OK",
	"201": "Created",
	"202": "Accepted",
	"203": "Non-Authoritative Information",
	"204": "No Content",
	"205": "Reset Content",
	"206": "Partial Content",
	"207": "Multi-Status (WebDAV)",
	"208": "Already Reported (WebDAV)",
	"226": "IM Used (HTTP Delta encoding)",
	"300": "Multiple Choices",
	"301": "Moved Permanently",
	"302": "Found",
	"303": "See Other",
	"304": "Not Modified",
	"305": "Use Proxy Deprecated",
	"306": "unused",
	"307": "Temporary Redirect",
	"308": "Permanent Redirect",
	"400": "Bad Request",
	"401": "Unauthorized",
	"402": "Payment Required Experimental",
	"403": "Forbidden",
	"404": "Not Found",
	"405": "Method Not Allowed",
	"406": "Not Acceptable",
	"407": "Proxy Authentication Required",
	"408": "Request Timeout",
	"409": "Conflict",
	"410": "Gone",
	"411": "Length Required",
	"412": "Precondition Failed",
	"413": "Payload Too Large",
	"414": "URI Too Long",
	"415": "Unsupported Media Type",
	"416": "Range Not Satisfiable",
	"417": "Expectation Failed",
	"418": "I'm a teapot",
	"421": "Misdirected Request",
	"422": "Unprocessable Entity (WebDAV)",
	"423": "Locked (WebDAV)",
	"424": "Failed Dependency (WebDAV)",
	"425": "Too Early Experimental",
	"426": "Upgrade Required",
	"428": "Precondition Required",
	"429": "Too Many Requests",
	"431": "Request Header Fields Too Large",
	"451": "Unavailable For Legal Reasons",
	"500": "Internal Server Error",
	"501": "Not Implemented",
	"502": "Bad Gateway",
	"503": "Service Unavailable",
	"504": "Gateway Timeout",
	"505": "HTTP Version Not Supported",
	"506": "Variant Also Negotiates",
	"507": "Insufficient Storage (WebDAV)",
	"508": "Loop Detected (WebDAV)",
	"510": "Not Extended",
	"511": "Network Authentication Required",
}

type Code int

func (c Code) String() string {
	if msg, ok := customCodeMessages[c]; ok {
		return msg
	}
	code := strconv.Itoa(int(c))
	if msg, ok := httpStatusMessages[code]; ok {
		return msg
	}
	if len(code) >= 3 {
		if msg, ok := httpStatusMessages[code[:3]]; ok {
			return msg
		}
	}
	return "Code(" + code + ")"
}

type CError struct {
	Code            Code
	Err             error
	Message         string
	OriginalMessage string
}

func (e *CError) Error() string {
	return e.Message
}

// Format parse APIError to string with suitable format
// then write it to provided writer
func (e *CError) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('#') || st.Flag('+'):
			_, _ = fmt.Fprintf(st, "\ncode=%v message=%v", e.Code, e.Message)
			if e.OriginalMessage != "" {
				_, _ = fmt.Fprintf(st, " original=%s", e.OriginalMessage)
			}
			if e.Err != nil {
				_, _ = fmt.Fprintf(st, " cause=%+v", e.Err)
			}
			fallthrough
		default:
			_, _ = io.WriteString(st, e.Error())
		}
	case 's':
		_, _ = io.WriteString(st, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(st, "%q", e.Error())
	}
}

// MarshalJSON jsonize APIError to bytes
func (e *CError) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}

	b := make([]byte, 0, 2048)

	b = append(b, '{')
	b = append(b, `"code":`...)
	b = append(b, strconv.FormatInt(int64(e.Code), 10)...)

	if e.Err != nil {
		b = append(b, ',')
		b = append(b, `"err":`...)
		b = append(b, marshal(e.Err.Error())...)
	}

	b = append(b, ',')
	b = append(b, `"msg":`...)
	b = append(b, marshal(e.Message)...)

	if e.OriginalMessage != "" {
		b = append(b, ',')
		b = append(b, `"orig":`...)
		b = append(b, marshal(e.OriginalMessage)...)
	}

	b = append(b, ',')

	b = append(b, '}')
	return b, nil
}

func marshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		data, _ = json.Marshal(err)
	}
	return data
}
