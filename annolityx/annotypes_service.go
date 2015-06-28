package annolityx

import (
	"fmt"
	"github.com/euforia/simplelog"
	"github.com/metrilyx/annolityx/annolityx/annotations"
	"net/http"
	"strings"
)

type AnnoTypeHandle struct {
	Typestore annotations.IEventAnnotationTypes
	Path      string // holds endpoint path
	logger    *simplelog.Logger
}

func NewAnnoTypeHandle(path string, typestore annotations.IEventAnnotationTypes, logger *simplelog.Logger) *AnnoTypeHandle {
	return &AnnoTypeHandle{Typestore: typestore, Path: path, logger: GetLogger(logger)}
}

func (ath *AnnoTypeHandle) RegisterHandle() {
	http.Handle(ath.Path, ath)
}

func (ath *AnnoTypeHandle) GetHandler(r *http.Request) (interface{}, int) {
	var (
		rslt       interface{}
		err        error
		reqSubPath = strings.Replace(r.URL.Path, ath.Path, "", -1)
	)

	if reqSubPath == "" {
		rslt = ath.Typestore.ListTypes()
	} else {
		if rslt, err = ath.Typestore.GetType(reqSubPath); err != nil {
			if strings.HasPrefix(err.Error(), "Type not found") {
				return fmt.Sprintf(`{"error": "Not found: %s"}`, reqSubPath), 404
			}
			return fmt.Sprintf(`{"error": "%s"}`, err.Error()), 400
		}
	}
	return rslt, 200
}

func (ath *AnnoTypeHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		resp interface{}
		code int
	)

	switch r.Method {
	case "GET":
		resp, code = ath.GetHandler(r)
		break
	default:
		resp = map[string]string{
			"error": fmt.Sprintf("Method not supported: %s", r.Method)}
		code = 405
		break
	}

	WriteJsonResponse(w, r, resp, code)
	ath.logger.Info.Printf("%s %d %s\n", r.Method, code, r.URL.RequestURI())
}
