package parsers

import (
	"testing"
)

func Test_ParseTimeToEpoch(t *testing.T) {

	val, err := ParseTimeToEpoch("1h-ago")
	if err != nil {
		t.Errorf("%s %s", val, err)
		t.FailNow()
	}
	t.Logf("ParseTimeToEpoch('1h-ago') %f\n", val)

	val, err = ParseTimeToEpoch("2013.10.21-09:32:23")
	if err != nil {
		t.Errorf("%s %s", val, err)
		t.FailNow()
	}
	t.Logf("ParseTimeToEpoch('2013.10.21-09:32:23') %f\n", val)
}
