package handler

import "testing"

func TestRefreshAllRunGuard(t *testing.T) {
	h := &Handler{}

	if !h.tryStartRefreshAll() {
		t.Fatal("expected first refresh-all start to succeed")
	}
	if h.tryStartRefreshAll() {
		t.Fatal("expected second refresh-all start to be rejected")
	}

	h.finishRefreshAll()

	if !h.tryStartRefreshAll() {
		t.Fatal("expected refresh-all to be allowed after finish")
	}
}
