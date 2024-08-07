package spot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMinQuantity(t *testing.T) {
	assert.Equal(t, 1.5, GetMinQuantity("BTCUSDT", 48000, "s1", 1))
	assert.Equal(t, 1.5, GetMinQuantity("BTCUSDT", 0.001, "s1", 1))
	assert.Equal(t, 1.5, GetMinQuantity("BTCUSDT", 48000, "s1", 0))
	assert.Equal(t, 1.5, GetMinQuantity("BTCUSDT", 48000, "s1", 100))
	assert.Equal(t, 1.5, GetMinQuantity("BTCUSDT", 48000, "s1", 1000))
	assert.Equal(t, 1.5, GetMinQuantity("BTCUSDT", 48000, "s1", 10000))
}
