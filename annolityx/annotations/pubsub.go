package annotations

import (
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq3"
	"sort"
	"strings"
)

type IEventAnnotationPublisher interface {
	Publish(string, interface{}) error
}

func GetZmqTypeFromString(t string) (zmq.Type, error) {
	switch t {
	case "PUB":
		return zmq.PUB, nil
	case "SUB":
		return zmq.SUB, nil
	default:
		break
	}
	return zmq.PUB, fmt.Errorf("invalid type: %s", t)
}

type EventAnnoPublisher struct {
	zsock *zmq.Socket
}

func NewEventAnnoPublisher(listenAddr, zTypeStr string) (*EventAnnoPublisher, error) {
	p := EventAnnoPublisher{}

	zType, err := GetZmqTypeFromString(zTypeStr)
	if err != nil {
		return &p, err
	}

	sock, _ := zmq.NewSocket(zType)
	err = sock.Bind(listenAddr)
	if err != nil {
		return &p, err
	}
	p.zsock = sock

	return &p, nil
}

func (p *EventAnnoPublisher) Publish(pubId string, data interface{}) error {
	jbytes, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	p.zsock.Send(fmt.Sprintf("%s %s", strings.ToLower(pubId), jbytes), 0)
	return nil
}

type EventAnnoSubMessage struct {
	SubscriptionId string
	Data           string
}

type EventAnnoSubscriber struct {
	zsock *zmq.Socket
}

func NewEventAnnoSubscriber(connectUri string, typeStr string, subscriptionStrs []string) (*EventAnnoSubscriber, error) {
	var (
		err   error
		sock  *zmq.Socket
		zType zmq.Type
		s     = EventAnnoSubscriber{}
	)

	zType, err = GetZmqTypeFromString(typeStr)
	if err != nil {
		return &s, err
	}

	sock, err = zmq.NewSocket(zType)
	if err != nil {
		return &s, err
	}

	err = sock.Connect(connectUri)
	if err != nil {
		return &s, err
	}
	// Set subscriptions
	for _, v := range subscriptionStrs {
		if err = sock.SetSubscribe(v); err != nil {
			return &s, err
		}
	}
	s.zsock = sock
	return &s, nil
}

func (e *EventAnnoSubscriber) Close() {
	e.zsock.Close()
}

func (e *EventAnnoSubscriber) Recieve() (EventAnnoSubMessage, error) {
	evtAnnoMsg := EventAnnoSubMessage{}

	msg, err := e.zsock.Recv(0)
	if err != nil {
		return evtAnnoMsg, err
	}

	envMsg := strings.Split(msg, " ")
	if len(envMsg) < 2 {
		return evtAnnoMsg, fmt.Errorf("invalid message: %s", msg)
	}

	data := strings.Join(envMsg[1:], " ")

	evtAnnoMsg.SubscriptionId = envMsg[0]
	evtAnnoMsg.Data = data

	return evtAnnoMsg, nil
}

/*
 * Creates a hash from annotation types and tags
 *
 * Return:
 * 		string typeA,typeB{tagA:valA,tagB:valB}
 *
 */
func SubscriptionHash(types []string, tags map[string]string) (string, error) {
	if len(types) < 1 && len(tags) < 1 {
		return "", nil
	}

	sort.Strings(types)
	typesStr := strings.Join(types, ",")

	tagkeys := make([]string, len(tags))
	i := 0
	for k, _ := range tags {
		tagkeys[i] = k
		i++
	}
	if len(tags) > 0 {
		tagsStr := ""
		sort.Strings(tagkeys)
		for _, v := range tagkeys {
			tagsStr = fmt.Sprintf("%s%s:%s,", tagsStr, v, tags[v])
		}
		return fmt.Sprintf("%s{%s}", typesStr, tagsStr[:len(tagsStr)-1]), nil
	} else {
		return fmt.Sprintf("%s{}", typesStr), nil
	}
}
