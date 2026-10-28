package ui

import (
	"testing"

	"github.com/ilaziness/apidoc/internal/protocol/httpapi"
)

func TestKVRoundTrip(t *testing.T) {
	in := []httpapi.KV{
		{Key: "a", Value: "1", Enabled: true},
		{Key: "b", Value: "2", Enabled: false},
	}
	got := decodeKV(encodeKV(in))
	if len(got) != 2 || got[0].Key != "a" || got[1].Enabled {
		t.Fatalf("%+v", got)
	}
}

func TestMapRoundTrip(t *testing.T) {
	got := decodeMap(encodeMap(map[string]string{"baseUrl": "http://localhost"}))
	if got["baseUrl"] != "http://localhost" {
		t.Fatalf("%v", got)
	}
}
