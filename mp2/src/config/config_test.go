package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	adds, _ := IntroducerIPAddresses()

	for _, addr := range adds {
		t.Log("Introducer:", addr)
	}
	port, _ := Port()

	t.Log("Introducer port:", port)
}
