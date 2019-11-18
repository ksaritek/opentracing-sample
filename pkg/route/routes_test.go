package route_test

import (
	"github.com/ksaritek/opentracing-sample/pkg/route"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBirthDayMessage(t *testing.T) {
	layout := "2006-01-02"
	now, err := time.Parse(layout, "2001-05-01")
	if err != nil {
		t.Error(err)
		return
	}
	msg, err := route.MsgOfBirthday("dummy", "2001-05-01", now)

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, msg, "Hello, dummy! Happy birthday")
}

func TestPassedBirthDayMessage(t *testing.T) {
	layout := "2006-01-02"
	now, err := time.Parse(layout, "2001-05-01")
	if err != nil {
		t.Error(err)
		return
	}
	msg, err := route.MsgOfBirthday("dummy", "2001-04-01", now)

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, msg, "Hello, dummy! Your birthday is passed this year")
}

func TestComingBirthDayMessage(t *testing.T) {
	layout := "2006-01-02"
	now, err := time.Parse(layout, "2001-05-01")
	if err != nil {
		t.Error(err)
		return
	}
	msg, err := route.MsgOfBirthday("dummy", "2001-05-02", now)

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, msg, "Hello, dummy! Your birthday is in 1 day")
}

func TestBirthDayMessageInvalidDateFormat(t *testing.T) {
	layout := "2006-01-02"
	now, err := time.Parse(layout, "2001-05-01")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = route.MsgOfBirthday("dummy", "2001-5-02", now)

	if err != nil {
		//expected error message
		return
	}

	t.Error("2001-5-02 is invalid date format, must give error")
}
