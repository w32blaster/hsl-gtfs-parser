package main

import (
	"testing"

	"github.com/patrickbr/gtfsparser/gtfs"
	"github.com/stretchr/testify/assert"
)

func TestTimeConvertion(t *testing.T) {

	// Given:
	time := gtfs.Time{Hour: 18, Minute: 0, Second: 0}

	// When:
	intTime := extractTime(&time)

	// Then:
	assert.Equal(t, 1800, intTime)
}

func TestTimeConvertionIgnoreSecods(t *testing.T) {

	// Given:
	time := gtfs.Time{Hour: 9, Minute: 59, Second: 45}

	// When:
	intTime := extractTime(&time)

	// Then:
	assert.Equal(t, 959, intTime)
}
