package utils

import "golang.org/x/sync/singleflight"

type (
	SingleFlightGroup singleflight.Group
	callback          func() (interface{}, error)
)

func (g *SingleFlightGroup) Do(key string, callback callback) (interface{}, error) {
	value, err := g.Do(key, func() (interface{}, error) {
		return callback()
	})
	return value, err
}
