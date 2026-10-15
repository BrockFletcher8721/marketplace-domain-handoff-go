package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type OnboardRequest struct {
	Seller string `json:"seller"`
	Domain string `json:"domain"`
	Buyer  string `json:"buyer"`
}
type OnboardResult struct {
	Seller      string `json:"seller"`
	ZoneID      string `json:"zone_id"`
	Status      string `json:"status"`
	BuyerUpdate string `json:"buyer_update"`
}

func onboard(ctx context.Context, client *InfraiClient, in OnboardRequest) (OnboardResult, error) {
	d, err := client.AddDomain(ctx, in.Domain)
	if err != nil {
		return OnboardResult{}, err
	}
	if err = client.UpsertTXT(ctx, d.ZoneID, "_marketplace", "marketplace-verification="+in.Seller); err != nil {
		return OnboardResult{}, err
	}
	if err = client.VerifyDomain(ctx, in.Domain); err != nil {
		return OnboardResult{}, err
	}
	return OnboardResult{Seller: in.Seller, ZoneID: d.ZoneID, Status: "verification_pending", BuyerUpdate: "domain handoff started for " + in.Buyer}, nil
}

func handler(client *InfraiClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/onboard" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var in OnboardRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Domain) == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		result, err := onboard(r.Context(), client, in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
}
