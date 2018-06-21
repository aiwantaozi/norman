package writer

import (
	"io"
	"time"

	"github.com/rancher/norman/types"

	"github.com/pborman/uuid"
	k8stypes "k8s.io/apimachinery/pkg/types"
	utilnet "k8s.io/apimachinery/pkg/util/net"
)

const (
	levelMetadata = iota
	levelRequest
	levelRequestResponse
)

var (
	auditLevel = map[int]string{
		levelMetadata:        "MetadataLevel",
		levelRequest:         "RequestLevel",
		levelRequestResponse: "RequestResponseLevel",
	}
)

type AuditLog struct {
	AuditID                  k8stypes.UID `json:"auditID,omitempty"`
	RequestReceivedTimestamp time.Time    `json:"requestReceivedTimestamp,omitempty"`
	RequestURI               string       `json:"requestURI,omitempty"`
	RequestBody              interface{}  `json:"requestBody,omitempty"`
	ResponseBody             interface{}  `json:"responseBody,omitempty"`
	ResponseStatus           int          `json:"responseStatus,omitempty"`
	SourceIPs                []string     `json:"sourceIPs,omitempty"`
	User                     *userInfo    `json:"user,omitempty"`
	UserAgent                string       `json:"userAgent,omitempty"`
	Verb                     string       `json:"verb,omitempty"`
	Level                    string       `json:"level,omitempty"`
}

type userInfo struct {
	Name  string `json:"name,omitempty"`
	Group string `json:"group,omitempty"`
}

type AuditLogWriter struct {
	Level   int
	Output  io.Writer
	Encoder func(io.Writer, interface{}) error
}

func (a *AuditLogWriter) Write(apiContext *types.APIContext) error {
	al := &AuditLog{
		AuditID: k8stypes.UID(uuid.NewRandom().String()),
		Level:   auditLevel[a.Level],
		RequestReceivedTimestamp: time.Now(),
		RequestURI:               apiContext.Request.RequestURI,
		ResponseStatus:           apiContext.ResponseStatus,
		Verb:                     apiContext.Request.Method,
	}

	ips := utilnet.SourceIPs(apiContext.Request)
	sourceIPs := make([]string, len(ips))
	for i := range ips {
		sourceIPs[i] = ips[i].String()
	}
	al.SourceIPs = sourceIPs

	al.User = &userInfo{
		Name:  apiContext.Request.Header.Get("Impersonate-User"),
		Group: apiContext.Request.Header.Get("Impersonate-Group"),
	}

	if a.Level >= levelRequest {
		al.RequestBody = apiContext.RequestBody
	}

	if a.Level >= levelRequestResponse {
		al.ResponseBody = apiContext.ResponseBody
	}

	return a.Encoder(a.Output, al)
}
