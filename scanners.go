package sql

import (
	"strings"
	"time"

	"github.com/gildas/go-errors"
)

type DBTime time.Time

func (t *DBTime) Scan(blob interface{}) (err error) {
	var parsed time.Time

	if payload, ok := blob.(time.Time); ok {
		parsed = payload
	} else if payload, ok := blob.([]byte); ok {
		value := strings.Split(string(payload), " m=")[0] // remove the monotonic clock if present as GO cannot parse it
		if parsed, err = time.Parse("2006-01-02 15:04:05 -0700 MST", value); err != nil {
			return errors.Unsupported.Wrap(err)
		}
	}
	*t = DBTime(parsed)
	return nil
}