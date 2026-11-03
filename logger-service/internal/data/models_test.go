package data

import (
	"context"
	"errors"
	"testing"
)

type mockInserter struct {
	last LogEntry
	err  error
}

func (m *mockInserter) Insert(ctx context.Context, entry LogEntry) error {
	m.last = entry
	return m.err
}

func TestInsertViaInterface(t *testing.T) {
	mock := &mockInserter{}
	var inserter EntryInserter = mock

	err := inserter.Insert(context.Background(), LogEntry{Name: "auth", Data: "ok"})
	if err != nil {
		t.Fatal(err)
	}
	if mock.last.Name != "auth" {
		t.Fatalf("expected name auth, got %s", mock.last.Name)
	}

	mock.err = errors.New("fail")
	if err := inserter.Insert(context.Background(), LogEntry{}); err == nil {
		t.Fatal("expected error")
	}
}
