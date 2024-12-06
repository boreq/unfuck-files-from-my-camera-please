package extractor

import (
	"time"

	"github.com/boreq/errors"
)

type Info struct {
	timestamp time.Time
}

func NewInfo(timestamp time.Time) (Info, error) {
	if time.Now().IsZero() {
		return Info{}, errors.New("timestamp is a zero value which implies that something went very wrong")
	}
	return Info{
		timestamp: timestamp,
	}, nil
}

func MustNewInfo(timestamp time.Time) Info {
	v, err := NewInfo(timestamp)
	if err != nil {
		panic(err)
	}
	return v
}

func (m *Info) Timestmap() time.Time {
	return m.timestamp
}
