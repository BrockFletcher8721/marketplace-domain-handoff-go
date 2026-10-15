package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOnboardUsesZoneAndReturnsBuyerHandoff(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{"zone_id":"z_123","domain":"shop.example"}}`))
	}))
	defer srv.Close()
	c := &InfraiClient{baseURL: srv.URL, key: "test", httpClient: srv.Client()}
	got, err := onboard(context.Background(), c, OnboardRequest{Seller: "seller-7", Domain: "shop.example", Buyer: "buyer-2"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ZoneID != "z_123" || got.Status != "verification_pending" || got.BuyerUpdate == "" {
		t.Fatalf("unexpected result: %+v", got)
	}
}
