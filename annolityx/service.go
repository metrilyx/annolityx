package annolityx

import (
	"encoding/json"
	"fmt"
	"github.com/euforia/simplelog"
	"github.com/metrilyx/annolityx/annolityx/annotations"
	"github.com/metrilyx/annolityx/annolityx/config"
	"github.com/metrilyx/annolityx/annolityx/datastores"
	"github.com/metrilyx/annolityx/annolityx/parsers"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const ACL_DEFAULT_ORIGIN string = "*"

type ServiceEndpoints struct {
	wsock string
	anno  string
	types string
}

type EventAnnoService struct {
	ListenAddr string
	Webroot    string
	Endpoints  ServiceEndpoints

	Typestore annotations.IEventAnnotationTypes
	Datastore annotations.IEventAnnotation
	Publisher annotations.IEventAnnotationPublisher

	pubSubPort int64
	cfg        *config.Config
	logger     *simplelog.Logger

	wsClients int64
}

func NewEventAnnoService(cfg *config.Config, logger *simplelog.Logger) (*EventAnnoService, error) {
	var err error

	eas := EventAnnoService{
		Webroot:    cfg.Http.Webroot,
		ListenAddr: fmt.Sprintf(":%d", cfg.Http.Port),
		pubSubPort: cfg.Publisher.Port,
		cfg:        cfg,
		wsClients:  0,
	}

	eas.logger = GetLogger(logger)

	eas.Endpoints = ServiceEndpoints{
		cfg.Http.WebsocketEndpoint,
		cfg.Http.AnnoEndpoint,
		cfg.Http.TypesEndpoint,
	}

	if eas.Typestore, err = datastores.NewJsonFileTypestore(cfg.Typestore.DBFile); err != nil {
		return &eas, err
	}

	if eas.Datastore, err = datastores.NewElasticsearchDatastore(cfg); err != nil {
		return &eas, err
	}

	pubAddr := fmt.Sprintf("tcp://*:%d", cfg.Publisher.Port)
	pub, err := annotations.NewEventAnnoPublisher(pubAddr, cfg.Publisher.Type)
	if err != nil {
		return &eas, err
	}
	eas.Publisher = pub
	logger.Warning.Printf("Publisher started on: %s\n", pubAddr)

	return &eas, nil
}

func (e *EventAnnoService) Start() error {
	e.logger.Warning.Printf("HTTP root directory: %s\n", e.Webroot)
	http.Handle("/", http.FileServer(http.Dir(e.Webroot)))

	e.logger.Warning.Printf("Registering websocket endpoint: %s\n", e.Endpoints.wsock)
	wsHdl := NewWebSockService(e.Endpoints.wsock, e.pubSubPort, e.logger)
	wsHdl.RegisterHandle()

	e.logger.Warning.Printf("Registering config endpoint: /api/config\n")
	cfgHdl := NewConfigHandle("/api/config", e.cfg, e.logger)
	cfgHdl.RegisterHandle()

	e.logger.Warning.Printf("Registering types endpoint: %s\n", e.Endpoints.types)
	typeHdl := NewAnnoTypeHandle(e.Endpoints.types, e.Typestore, e.logger)
	typeHdl.RegisterHandle()

	e.logger.Warning.Printf("Registering annotation endpoint: %s\n", e.Endpoints.anno)
	http.HandleFunc(e.Endpoints.anno, e.annotationHandler)

	if strings.HasSuffix(e.Endpoints.anno, "/") {
		http.HandleFunc(e.Endpoints.anno[:len(e.Endpoints.anno)-1], e.annotationHandler)
	} else {
		http.HandleFunc(fmt.Sprintf("%s/", e.Endpoints.anno), e.annotationHandler)
	}

	e.logger.Warning.Printf("Starting HTTP service %s...\n", e.ListenAddr)
	return http.ListenAndServe(e.ListenAddr, nil)
}

func (e *EventAnnoService) checkAnnotateRequest(r *http.Request) (*annotations.EventAnnotation, error) {
	var annoReq annotations.EventAnnotation

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &annoReq, err
	}

	err = json.Unmarshal(body, &annoReq)
	if err != nil {
		return &annoReq, err
	}

	annoReq.Type = strings.ToLower(annoReq.Type)

	_, err = e.Typestore.GetType(annoReq.Type)
	if err != nil {
		return &annoReq, err
	}

	if annoReq.Message == "" || len(annoReq.Tags) < 1 {
		return &annoReq, fmt.Errorf("Missing 'type', 'message', or 'tags'!")
	}

	if annoReq.Timestamp == 0 {
		annoReq.Timestamp = float64(time.Now().UnixNano()) / 1000000000
	}
	return &annoReq, nil
}

func (e *EventAnnoService) parseRequestPath(r *http.Request) []string {
	parts := make([]string, 0)
	for _, s := range strings.Split(r.URL.Path, "/") {
		if s != "" {
			parts = append(parts, s)
		}
	}
	return parts
}

func (e *EventAnnoService) handleAnnoGetRequest(r *http.Request) (interface{}, int) {
	reqPathParts := e.parseRequestPath(r)

	if len(reqPathParts) == 4 {
		resp, err := e.Datastore.Get(reqPathParts[2], reqPathParts[3])
		if err != nil {
			if err.Error() == "record not found" {
				return map[string]string{"error": err.Error()}, 404
			}
			return map[string]string{"error": err.Error()}, 400
		}
		return resp, 200
	}

	pp := parsers.AnnoQueryParamsParser{r.URL.Query(), r.Body}
	q, err := pp.ParseGetParams()
	if err != nil {
		return err.Error(), 400
	} else {
		var limit int64
		if val, ok := r.URL.Query()["limit"]; ok {
			if limit, err = strconv.ParseInt(val[0], 10, 64); err != nil {
				return fmt.Sprintf(`{"error": "%s"}`, err), 400
			}
		} else {
			limit = 0
		}
		rslt, err := e.Datastore.Query(*q, limit)
		if err != nil {
			return fmt.Sprintf(`{"error": "%s"}`, err), 401
		}
		return rslt, 200
	}
}

func (e *EventAnnoService) handleAnnoPostPutRequest(r *http.Request) (interface{}, int) {
	evtAnno, err := e.checkAnnotateRequest(r)
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error()), 400
	}
	resp, err := e.Datastore.Annotate(evtAnno)
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error()), 401
	}

	if err = e.Publisher.Publish(resp.Type, resp); err != nil {
		e.logger.Warning.Printf("Failed to publish: %s\n", err)
	} else {
		e.logger.Trace.Printf("Published (%s): %#v\n", resp.Type, resp)
	}
	return resp, 200
}

func (e *EventAnnoService) annotationHandler(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	var code int

	switch r.Method {
	case "GET":
		resp, code = e.handleAnnoGetRequest(r)
		break
	case "POST", "PUT":
		resp, code = e.handleAnnoPostPutRequest(r)
		break
	default:
		e.logger.Trace.Printf(`{"error": "Method not supported: %s"}`, r.Method)
		resp = map[string]string{"error": fmt.Sprintf("Method not supported: %s", r.Method)}
		code = 501
		break
	}
	WriteJsonResponse(w, r, resp, code)
	e.logger.Info.Printf("%s %d %s\n", r.Method, code, r.URL.RequestURI())
}

func WriteJsonResponse(w http.ResponseWriter, r *http.Request, data interface{}, respCode int) {
	var b []byte
	s, ok := data.(string)
	if ok {
		b = []byte(s)
	} else {
		b, _ = json.Marshal(&data)
	}

	w.Header().Set("Access-Control-Allow-Origin", ACL_DEFAULT_ORIGIN)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(respCode)
	w.Write(b)
}

func GetLogger(logger *simplelog.Logger) *simplelog.Logger {
	if logger == nil {
		return simplelog.NewStdLogger()
	}
	return logger
}
