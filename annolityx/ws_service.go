package annolityx

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"github.com/euforia/simplelog"
	"github.com/metrilyx/annolityx/annolityx/annotations"
	"net/http"
	"time"
)

type WebSockService struct {
	Path       string
	wsClients  int64
	pubSubPort int64
	logger     *simplelog.Logger
}

func NewWebSockService(path string, pubsubport int64, logger *simplelog.Logger) *WebSockService {
	return &WebSockService{
		Path:       path,
		wsClients:  0,
		logger:     GetLogger(logger),
		pubSubPort: pubsubport,
	}
}

func (w *WebSockService) RegisterHandle() {
	http.Handle(w.Path, websocket.Handler(w.WebsockHandler))
}

func (w *WebSockService) getSubscription(ws *websocket.Conn) (annotations.Subscription, error) {

	var (
		subMsg annotations.Subscription
		err    error
	)

	var rawData []byte
	if err = websocket.Message.Receive(ws, &rawData); err != nil {
		return subMsg, err
	}
	//e.logger.Trace.Printf("Subscription message: %s\n", rawData)

	if err = json.Unmarshal(rawData, &subMsg); err != nil {
		return subMsg, fmt.Errorf("Invalid subscription id: %s %s", rawData, err)
	}

	return subMsg, nil
}

func (w *WebSockService) SubcriptionURI() string {
	return fmt.Sprintf("tcp://localhost:%d", w.pubSubPort)
}

func (w *WebSockService) WebsockHandler(ws *websocket.Conn) {
	var (
		err                error
		clientSubscription annotations.Subscription
		subscriber         *annotations.EventAnnoSubscriber
	)

	w.wsClients++
	w.logger.Info.Printf("WebSocket client connected: %s (clients: %d)\n",
		ws.Request().RemoteAddr, w.wsClients)

	if clientSubscription, err = w.getSubscription(ws); err != nil {
		w.logger.Error.Printf("Failed to get subscriptions: %s\n", err)

		websocket.Message.Send(ws, fmt.Sprintf(`{"error": "%s"}`, err))

		w.wsClients--
		return
	}
	w.logger.Info.Printf("Subscription (%s): '%s'\n", ws.Request().RemoteAddr, clientSubscription)

	if subscriber, err = annotations.NewEventAnnoSubscriber(w.SubcriptionURI(),
		"SUB", clientSubscription.Types); err != nil {

		w.logger.Error.Printf("Failed to start subscriber: %s", err)
		websocket.Message.Send(ws,
			fmt.Sprintf(`{"error": "Failed to start subscriber: %s"}`, err.Error()))

		w.wsClients--
		return
	}
	w.logger.Debug.Printf("Client subscriber connected: %s\n", w.SubcriptionURI())

	// Precautionary - might be able to remove.
	defer subscriber.Close()

	// Holder for client disconnect detection.
	//var tmpd string
	for {
		// Check for client disconnect.
		/*if err = websocket.Message.Receive(ws, &tmpd); err != nil {

			if err = subscriber.Close(); err != nil {
				w.logger.Error.Printf("Could not close subscriber: %s\n", err)
			}
			w.wsClients--

			w.logger.Warning.Printf("Client disconnected: %s (clients: %d)\n",
				ws.Request().RemoteAddr, w.wsClients)
			return
		}*/

		//w.logger.Trace.Printf("Waiting for subscription message...\n")
		evtAnnoMsg, err := subscriber.Receive()
		if err != nil {
			w.logger.Error.Printf("Failed to recieve subscription message: %s\n", err)
			continue
		}
		w.logger.Trace.Printf("Subscription message recieved: %s", evtAnnoMsg)

		var annoCfm annotations.EventAnnotation
		err = json.Unmarshal([]byte(evtAnnoMsg.Data), &annoCfm)
		if err != nil {
			w.logger.Error.Printf("Decode failure: %s\n", evtAnnoMsg)
			continue
		}

		if clientSubscription.IsSubscribedMessage(annoCfm) {
			/* publish (send data over websocket) only if the timestamp is within the last minute */
			if annoCfm.Timestamp > float64(time.Now().Unix()-60) {
				websocket.Message.Send(ws, evtAnnoMsg.Data)
			} else {
				w.logger.Trace.Printf("Retro-active posting: %s\n", evtAnnoMsg.Data)
			}
		} else {
			w.logger.Trace.Printf("Message not subscribed :%s", annoCfm)
		}
	}
}
