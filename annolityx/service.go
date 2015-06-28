package annolityx

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"github.com/euforia/simplelog"
	"github.com/metrilyx/annolityx/annolityx/annotations"
	"github.com/metrilyx/annolityx/annolityx/config"
	"github.com/metrilyx/annolityx/annolityx/datastores"
	"github.com/metrilyx/annolityx/annolityx/parsers"
	"io/ioutil"
	"net/http"
	"os"
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
	eas := EventAnnoService{
		Webroot:    cfg.Http.Webroot,
		ListenAddr: fmt.Sprintf(":%d", cfg.Http.Port),
		pubSubPort: cfg.Publisher.Port,
		cfg:        cfg,
		wsClients:  0,
	}
	if logger == nil {
		eas.logger = simplelog.NewLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		eas.logger = logger
	}

	eas.Endpoints = ServiceEndpoints{
		cfg.Http.WebsocketEndpoint,
		cfg.Http.AnnoEndpoint,
		cfg.Http.TypesEndpoint,
	}

	ts, err := datastores.NewJsonFileTypestore(cfg.Typestore.DBFile)
	if err != nil {
		return &eas, err
	}
	eas.Typestore = ts

	ds, err := datastores.NewElasticsearchDatastore(cfg)
	if err != nil {
		return &eas, err
	}
	eas.Datastore = ds

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
	http.Handle(e.Endpoints.wsock, websocket.Handler(e.wsHandler))

	e.logger.Warning.Printf("Registering config endpoint: /api/config\n")
	http.HandleFunc("/api/config", e.configHandler)

	e.logger.Warning.Printf("Registering types endpoint: %s\n", e.Endpoints.types)
	http.HandleFunc(e.Endpoints.types, e.typesHandler)

	e.logger.Warning.Printf("Registering annotation ndpoint: %s\n", e.Endpoints.anno)
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

func (e *EventAnnoService) handleConfigGetRequest(r *http.Request) (interface{}, int) {
	if e.cfg.Http.WebsocketHostname == "" {

		var err error
		e.cfg.Http.WebsocketHostname, err = os.Hostname()

		if err != nil {
			return err, 500
		}
	}

	var eConfig = map[string]interface{}{
		"websocket": map[string]string{
			"url": fmt.Sprintf(`ws://%s:%d%s`, e.cfg.Http.WebsocketHostname, e.cfg.Http.Port,
				e.cfg.Http.WebsocketEndpoint),
		},
		"endpoints": map[string]string{
			"types":      e.cfg.Http.TypesEndpoint,
			"annotation": e.cfg.Http.AnnoEndpoint,
		},
	}

	return eConfig, 200
}

func (e *EventAnnoService) configHandler(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	var code int
	switch r.Method {
	case "GET":
		resp, code = e.handleConfigGetRequest(r)
		break
	default:
		resp = map[string]string{
			"error": fmt.Sprintf("Method not supported: %s", r.Method)}
		code = 405
		break
	}
	e.writeJsonResponse(w, r, resp, code)
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

	if err := e.Publisher.Publish(resp.Type, resp); err != nil {
		e.logger.Warning.Printf("Failed to publish: %s\n", err)
	} else {
		e.logger.Trace.Printf("Published (%s): %s\n", resp.Type, resp)
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
	e.writeJsonResponse(w, r, resp, code)
}

func (e *EventAnnoService) handleAnnoTypesRequest(r *http.Request) (interface{}, int) {
	var rslt interface{}
	var err error
	reqSubPath := strings.Replace(r.URL.Path, e.Endpoints.types, "", -1)
	if reqSubPath == "" {
		rslt = e.Typestore.ListTypes()
	} else {
		rslt, err = e.Typestore.GetType(reqSubPath)
		if err != nil {
			if strings.HasPrefix(err.Error(), "Type not found") {
				return fmt.Sprintf(`{"error": "Not found: %s"}`, reqSubPath), 404
			}
			return fmt.Sprintf(`{"error": "%s"}`, err.Error()), 400
		}
	}
	return rslt, 200
}

func (e *EventAnnoService) typesHandler(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	var code int
	switch r.Method {
	case "GET":
		resp, code = e.handleAnnoTypesRequest(r)
		break
	default:
		resp = map[string]string{
			"error": fmt.Sprintf("Method not supported: %s", r.Method)}
		code = 501
		break
	}
	e.writeJsonResponse(w, r, resp, code)
}

func (e *EventAnnoService) getSubscription(ws *websocket.Conn) (annotations.Subscription, error) {
	var subMsg annotations.Subscription

	var rawData []byte
	err := websocket.Message.Receive(ws, &rawData)
	if err != nil {
		return subMsg, err
	}
	//e.logger.Trace.Printf("Subscription message: %s\n", rawData)

	if err := json.Unmarshal(rawData, &subMsg); err != nil {
		return subMsg, fmt.Errorf("Invalid subscription id: %s %s", rawData, err)
	}

	return subMsg, nil
}

func (e *EventAnnoService) SubcriptionURI() string {
	return fmt.Sprintf("tcp://localhost:%d", e.pubSubPort)
}

func (e *EventAnnoService) wsHandler(ws *websocket.Conn) {
	var (
		err                error
		clientSubscription annotations.Subscription
		subscriber         *annotations.EventAnnoSubscriber
	)

	e.wsClients++
	e.logger.Info.Printf("WebSocket client connected: %s (clients: %d)\n",
		ws.Request().RemoteAddr, e.wsClients)

	if clientSubscription, err = e.getSubscription(ws); err != nil {
		e.logger.Error.Printf("Failed to get subscriptions: %s\n", err)

		websocket.Message.Send(ws, fmt.Sprintf(`{"error": "%s"}`, err))

		e.wsClients--
		return
	}
	e.logger.Info.Printf("Subscription (%s): '%s'\n", ws.Request().RemoteAddr, clientSubscription)

	if subscriber, err = annotations.NewEventAnnoSubscriber(e.SubcriptionURI(),
		"SUB", clientSubscription.Types); err != nil {

		e.logger.Error.Printf("Failed to start subscriber: %s", err)
		websocket.Message.Send(ws,
			fmt.Sprintf(`{"error": "Failed to start subscriber: %s"}`, err.Error()))

		e.wsClients--
		return
	}

	// Precautionary - might be able to remove.
	defer subscriber.Close()

	// Holder for client disconnect detection.
	var tmpd string
	for {
		// Check for client disconnect.
		if err = websocket.Message.Receive(ws, &tmpd); err != nil {

			if err = subscriber.Close(); err != nil {
				e.logger.Error.Printf("Could not close subscriber: %s\n", err)
			}
			e.wsClients--

			e.logger.Warning.Printf("Client disconnected: %s (clients: %d)\n",
				ws.Request().RemoteAddr, e.wsClients)
			return
		}

		evtAnnoMsg, err := subscriber.Recieve()
		if err != nil {
			e.logger.Error.Printf("Failed to recieve subscription message: %s\n", err)
			continue
		}

		var annoCfm annotations.EventAnnotation
		err = json.Unmarshal([]byte(evtAnnoMsg.Data), &annoCfm)
		if err != nil {
			e.logger.Error.Printf("Decode failure: %s\n", evtAnnoMsg)
			continue
		}

		if clientSubscription.IsSubscribedMessage(annoCfm) {
			/* publish (send data over websocket) only if the timestamp is within the last minute */
			if annoCfm.Timestamp > float64(time.Now().Unix()-60) {
				websocket.Message.Send(ws, evtAnnoMsg.Data)
			} else {
				e.logger.Trace.Printf("Retro-active posting: %s\n", evtAnnoMsg.Data)
			}
		} else {
			e.logger.Trace.Printf("Message not subscribed :%s", annoCfm)
		}
	}
}

func (e *EventAnnoService) writeJsonResponse(w http.ResponseWriter, r *http.Request, data interface{}, respCode int) {
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
	e.logger.Info.Printf("%s %d %s\n", r.Method, respCode, r.URL.RequestURI())
}
