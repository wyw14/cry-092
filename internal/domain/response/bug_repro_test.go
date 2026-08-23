package response

import (
	"testing"
	"time"
)

func TestReplyEvidenceAndFirstViewRemainImmutable(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	files := []string{"formal-reply.pdf", "evidence.jpg"}
	reply, err := NewReply("r1", "p1", "officer", 1, KindResolved, "已完成现场整改", files, now)
	if err != nil {
		t.Fatal(err)
	}
	files[0] = "replaced-after-submit"
	if reply.FileIDs[0] != "formal-reply.pdf" || reply.FileIDs[1] != "evidence.jpg" {
		t.Fatalf("stored reply evidence was mutated: %v", reply.FileIDs)
	}
	firstView := now.Add(time.Hour)
	if !reply.MarkViewed(firstView) {
		t.Fatal("first view was not recorded")
	}
	if reply.MarkViewed(firstView.Add(time.Hour)) {
		t.Fatal("second view replaced the first-view record")
	}
	if !reply.ViewedAt.Equal(firstView) {
		t.Fatalf("first viewed time changed to %s", reply.ViewedAt)
	}
}
