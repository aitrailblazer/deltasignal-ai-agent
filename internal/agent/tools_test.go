package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type stubTripCodeResolver struct {
	resp TripCodeResearchResponse
	err  error
}

func (r stubTripCodeResolver) ResolveTripCodeResearchPacket(context.Context, TripCodeResearchRequest) (TripCodeResearchResponse, error) {
	return r.resp, r.err
}

func TestDemoToolClientAndNormalizeIssuer(t *testing.T) {
	client := DemoToolClient{}
	for _, call := range []struct {
		name string
		fn   func(context.Context, string) (SpecialistResult, error)
	}{
		{"stress-scanner", client.StressSignals},
		{"evidence-retriever", client.CompanyEvidence},
		{"peer-analyst", client.PeerContext},
	} {
		result, err := call.fn(context.Background(), " hut ")
		if err != nil {
			t.Fatalf("%s returned error: %v", call.name, err)
		}
		if result.Agent != call.name || !strings.Contains(result.Summary, "HUT") || len(result.Evidence) != 1 {
			t.Fatalf("unexpected %s result: %#v", call.name, result)
		}
	}
	if normalizeIssuer("") != "the issuer" || normalizeIssuer(" hut ") != "HUT" {
		t.Fatal("normalizeIssuer branches failed")
	}
}

func TestDemoTripCodeResolver(t *testing.T) {
	resolver := DemoTripCodeResolver{}
	resp, err := resolver.ResolveTripCodeResearchPacket(context.Background(), TripCodeResearchRequest{
		TripCode: " tf-sub-9da70a7f98 ",
		Issuer:   "hut",
	})
	if err != nil {
		t.Fatalf("ResolveTripCodeResearchPacket returned error: %v", err)
	}
	if resp.TripCode != "TF-SUB-9DA70A7F98" || resp.Issuer != "HUT" || resp.Mode != "demo-tripcode-river" {
		t.Fatalf("unexpected demo response: %#v", resp)
	}
	if !strings.Contains(resp.Packet["article"].(map[string]any)["title"].(string), "Hut 8") {
		t.Fatalf("expected HUT article title, got %#v", resp.Packet["article"])
	}
	if resp.Packet["river"].(map[string]any)["node_count"] != 10 {
		t.Fatalf("expected 10-node demo river, got %#v", resp.Packet["river"])
	}

	generic, err := resolver.ResolveTripCodeResearchPacket(context.Background(), TripCodeResearchRequest{TripCode: "TF-SUB-OTHER"})
	if err != nil {
		t.Fatalf("generic TripCode returned error: %v", err)
	}
	if generic.Issuer != "HUT" || generic.Packet["article"].(map[string]any)["title"] != "DeltaSignal TripCode Research Packet" {
		t.Fatalf("unexpected generic demo packet: %#v", generic)
	}

	if _, err := resolver.ResolveTripCodeResearchPacket(context.Background(), TripCodeResearchRequest{TripCode: "bad"}); err == nil {
		t.Fatal("expected invalid TripCode error")
	}
}

func TestFallbackTripCodeResolver(t *testing.T) {
	primaryResp := TripCodeResearchResponse{TripCode: "TF-SUB-PRIMARY", Packet: map[string]any{"source": "primary"}}
	resolver := FallbackTripCodeResolver{
		Primary:  stubTripCodeResolver{resp: primaryResp},
		Fallback: DemoTripCodeResolver{},
	}
	resp, err := resolver.ResolveTripCodeResearchPacket(context.Background(), TripCodeResearchRequest{TripCode: "TF-SUB-X"})
	if err != nil {
		t.Fatalf("primary resolver returned error: %v", err)
	}
	if resp.Packet["source"] != "primary" {
		t.Fatalf("expected primary response, got %#v", resp.Packet)
	}

	resolver.Primary = stubTripCodeResolver{err: errors.New("unsupported tool")}
	resp, err = resolver.ResolveTripCodeResearchPacket(context.Background(), TripCodeResearchRequest{TripCode: "TF-SUB-X"})
	if err != nil {
		t.Fatalf("fallback resolver returned error: %v", err)
	}
	if !strings.Contains(resp.Packet["live_mcp_error"].(string), "unsupported tool") {
		t.Fatalf("expected live_mcp_error in fallback packet, got %#v", resp.Packet)
	}

	resolver.Fallback = stubTripCodeResolver{resp: TripCodeResearchResponse{TripCode: "TF-SUB-NIL-PACKET"}}
	resp, err = resolver.ResolveTripCodeResearchPacket(context.Background(), TripCodeResearchRequest{TripCode: "TF-SUB-X"})
	if err != nil {
		t.Fatalf("nil-packet fallback resolver returned error: %v", err)
	}
	if !strings.Contains(resp.Packet["live_mcp_error"].(string), "unsupported tool") {
		t.Fatalf("expected live_mcp_error in initialized fallback packet, got %#v", resp.Packet)
	}

	resolver.Fallback = nil
	if _, err := resolver.ResolveTripCodeResearchPacket(context.Background(), TripCodeResearchRequest{TripCode: "TF-SUB-X"}); err == nil {
		t.Fatal("expected primary-only error")
	}

	resolver.Primary = nil
	if _, err := resolver.ResolveTripCodeResearchPacket(context.Background(), TripCodeResearchRequest{TripCode: "TF-SUB-X"}); err == nil {
		t.Fatal("expected unconfigured resolver error")
	}

	resolver.Fallback = DemoTripCodeResolver{}
	resp, err = resolver.ResolveTripCodeResearchPacket(context.Background(), TripCodeResearchRequest{TripCode: "TF-SUB-X"})
	if err != nil {
		t.Fatalf("fallback-only resolver returned error: %v", err)
	}
	if resp.Mode != "demo-tripcode-river" {
		t.Fatalf("expected fallback-only demo response, got %#v", resp)
	}

	resolver.Primary = stubTripCodeResolver{err: errors.New("primary failed")}
	resolver.Fallback = stubTripCodeResolver{err: errors.New("fallback failed")}
	if _, err := resolver.ResolveTripCodeResearchPacket(context.Background(), TripCodeResearchRequest{TripCode: "TF-SUB-X"}); err == nil || !strings.Contains(err.Error(), "fallback failed") {
		t.Fatalf("expected combined fallback error, got %v", err)
	}

	if firstNonEmpty("", " hut ") != " hut " || firstNonEmpty("", "") != "" {
		t.Fatal("firstNonEmpty branches failed")
	}
}
