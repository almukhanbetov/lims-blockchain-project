package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// postJSON sends body as JSON to the given router and returns the decoded response.
func postJSON(t *testing.T, r http.Handler, path string, body map[string]string) (int, map[string]interface{}) {
	t.Helper()

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response %s: %v", rec.Body.String(), err)
	}
	return rec.Code, resp
}

// TestRegisterAndVerify covers the H2O-0003 scenario: register an event, verify
// it with the exact same fields (must match), then verify again with a tampered
// result field (must not match). Runs with bc == nil, so this exercises the
// local-cache path deterministically without a live blockchain.
func TestRegisterAndVerify(t *testing.T) {
	r := newRouter()

	base := map[string]string{
		"lims_record_id": "H2O-0003",
		"event_type":     "RESULT_VERIFIED",
		"sample_id":      "H2O-0003",
		"result":         "pH=7.2",
		"user_id":        "lab_manager",
		"status":         "VERIFIED",
		"timestamp":      "2026-07-05T20:00:00Z",
	}

	status, hashResp := postJSON(t, r, "/events/hash", base)
	if status != http.StatusOK {
		t.Fatalf("POST /events/hash: expected 200, got %d (%v)", status, hashResp)
	}
	if hashResp["success"] != true {
		t.Fatalf("POST /events/hash: expected success=true, got %v", hashResp)
	}
	registeredHash, _ := hashResp["data_hash"].(string)
	if registeredHash == "" {
		t.Fatalf("POST /events/hash: expected non-empty data_hash, got %v", hashResp)
	}

	// Same fields must verify successfully.
	status, verifyResp := postJSON(t, r, "/events/verify", base)
	if status != http.StatusOK {
		t.Fatalf("POST /events/verify (matching): expected 200, got %d (%v)", status, verifyResp)
	}
	if verifyResp["verified"] != true {
		t.Fatalf("POST /events/verify (matching): expected verified=true, got %v", verifyResp)
	}
	if verifyResp["expected_hash"] != registeredHash || verifyResp["actual_hash"] != registeredHash {
		t.Fatalf("POST /events/verify (matching): hash mismatch in response: %v", verifyResp)
	}

	// Tampering the result field (pH=7.2 -> pH=8.4) must fail verification.
	tampered := map[string]string{}
	for k, v := range base {
		tampered[k] = v
	}
	tampered["result"] = "pH=8.4"

	status, tamperedResp := postJSON(t, r, "/events/verify", tampered)
	if status != http.StatusOK {
		t.Fatalf("POST /events/verify (tampered): expected 200, got %d (%v)", status, tamperedResp)
	}
	if tamperedResp["verified"] != false {
		t.Fatalf("POST /events/verify (tampered): expected verified=false, got %v", tamperedResp)
	}
	if tamperedResp["expected_hash"] != registeredHash {
		t.Fatalf("POST /events/verify (tampered): expected_hash should still be the original hash, got %v", tamperedResp)
	}
	if tamperedResp["actual_hash"] == registeredHash {
		t.Fatalf("POST /events/verify (tampered): actual_hash should differ from the original, got %v", tamperedResp)
	}
}
