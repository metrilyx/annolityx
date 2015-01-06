package annotations

import (
	"testing"
	"time"
)

var testEvtAnnoCfm EventAnnotation = EventAnnotation{
	Id:              "testid",
	Type:            "test",
	Message:         "Test message",
	Tags:            map[string]string{"test": "test"},
	Data:            map[string]interface{}{"contact": "test"},
	Timestamp:       float64(time.Now().UnixNano()) / 1000000000,
	PostedTimestamp: float64(time.Now().UnixNano()) / 1000000000,
}

func Test_Subscription(t *testing.T) {
	subscript := Subscription{make([]string, 0), map[string]string{"test": "test"}}

	if !subscript.IsSubscribedMessage(testEvtAnnoCfm) {
		t.Errorf("Subscription check failed!\n")
		t.FailNow()
	}
}
func Test_Subscription_NoTags(t *testing.T) {
	subscript := Subscription{make([]string, 0), make(map[string]string)}

	if !subscript.IsSubscribedMessage(testEvtAnnoCfm) {
		t.Errorf("Subscription check (no tags) failed!\n")
		t.FailNow()
	}
}
func Test_Subscription_NotSubscribed(t *testing.T) {
	subscript := Subscription{make([]string, 0), map[string]string{"test": "sam"}}

	if subscript.IsSubscribedMessage(testEvtAnnoCfm) {
		t.Errorf("Subscription check (not subscribed) failed!\n")
		t.FailNow()
	}
}
