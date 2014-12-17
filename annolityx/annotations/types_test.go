package annotations

import (
	"testing"
)

func Test_NewEventAnnoType(t *testing.T) {
	evtAnnoType := NewEventAnnoType("test", "Test")
	t.Logf("%s\n", evtAnnoType)
}
