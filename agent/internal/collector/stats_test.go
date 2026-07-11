package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCollectAliveIPs_ClashMetadata(t *testing.T) {
	payload := `{
		"connections": [
			{
				"id": "c1",
				"upload": 100,
				"download": 200,
				"metadata": {
					"sourceIP": "192.168.25.1",
					"type": "vless/vless-in"
				}
			},
			{
				"id": "c2",
				"upload": 50,
				"download": 80,
				"metadata": {
					"sourceIP": "10.0.0.2",
					"type": "vless/vless-in",
					"user": "4eadc575-8b60-4964-89e3-3bdf270b27e6"
				}
			}
		],
		"downloadTotal": 280,
		"uploadTotal": 150
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/connections" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	col := New(srv.URL, "1")
	uuid := "4eadc575-8b60-4964-89e3-3bdf270b27e6"
	col.SetKnownUsers([]string{uuid})

	alive, err := col.CollectAliveIPs()
	if err != nil {
		t.Fatal(err)
	}
	// single known user fallback + explicit user on c2 should map to same uuid
	ips := alive[uuid]
	if len(ips) != 2 {
		t.Fatalf("expected 2 unique IPs for user, got %v", alive)
	}

	// first collect establishes baselines; second should report delta
	// mutate payload counters
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		t.Fatal(err)
	}
	conns := m["connections"].([]interface{})
	c0 := conns[0].(map[string]interface{})
	c0["upload"] = float64(150)
	c0["download"] = float64(300)
	b, _ := json.Marshal(m)

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(b)
	}))
	defer srv2.Close()
	col2 := New(srv2.URL, "1")
	col2.SetKnownUsers([]string{uuid})
	// first snapshot
	if _, err := col2.Collect(); err != nil {
		t.Fatal(err)
	}
	// bump
	c0["upload"] = float64(200)
	c0["download"] = float64(400)
	b2, _ := json.Marshal(m)
	srv2.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(b2)
	})
	delta, err := col2.CollectXboard()
	if err != nil {
		t.Fatal(err)
	}
	d := delta[uuid]
	if d[0] != 50 || d[1] != 100 {
		t.Fatalf("expected delta up=50 down=100, got %v", d)
	}
}

func TestCollect_NoUserNoFallback(t *testing.T) {
	payload := `{"connections":[{"id":"x","upload":1,"download":2,"metadata":{"sourceIP":"1.2.3.4","type":"vless/vless-in"}}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()
	col := New(srv.URL, "1")
	// multiple known users => no fallback attribution
	col.SetKnownUsers([]string{"u1-uuid-aaaaaaaaaaaaaaaaaaaaaaaa", "u2-uuid-bbbbbbbbbbbbbbbbbbbbbbbb"})
	alive, err := col.CollectAliveIPs()
	if err != nil {
		t.Fatal(err)
	}
	if len(alive) != 0 {
		t.Fatalf("expected empty without user field when multi-user, got %v", alive)
	}
}
