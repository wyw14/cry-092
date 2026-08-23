package response

import (
	"testing"
	"time"
)

func TestNewReplyFileIDsDetachedFromCaller(t *testing.T) {
	files := []string{"f1", "f2", "f3"}
	r, err := NewReply("r", "p", "officer", 1, KindResolved, "done", files, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("NewReply: %v", err)
	}
	// 调用方在提交后改写、追加自己持有的附件数组，不得替换代表看到的正式答复文件证据。
	files[0] = "CHANGED"
	files = append(files, "f4")

	if len(r.FileIDs) != 3 || r.FileIDs[0] != "f1" || r.FileIDs[2] != "f3" {
		t.Fatalf("reply evidence mutated by caller: %v", r.FileIDs)
	}
}

func TestNewReplyCopiesIntoIndependentBuffer(t *testing.T) {
	files := []string{"f1", "f2", "f3"}
	r, err := NewReply("r", "p", "officer", 1, KindResolved, "done", files, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("NewReply: %v", err)
	}
	// 答复证据内部改动不得回写调用方原数组。
	r.FileIDs[0] = "MUTATED"
	if files[0] != "f1" {
		t.Fatalf("reply copy leaked into caller input: %v", files)
	}
}

func TestMarkViewedKeepsFirstReceiptTime(t *testing.T) {
	first := time.Date(2026, 8, 24, 1, 0, 0, 0, time.UTC)
	second := time.Date(2026, 8, 24, 2, 0, 0, 0, time.UTC)
	r, err := NewReply("r", "p", "officer", 1, KindResolved, "done", []string{"f"}, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("NewReply: %v", err)
	}

	if !r.MarkViewed(first) {
		t.Fatal("first MarkViewed should record the receipt")
	}
	if r.Version != 2 {
		t.Fatalf("version after first view = %d, want 2", r.Version)
	}
	// 代表第二次打开页面不得覆盖首次查收时间。
	if r.MarkViewed(second) {
		t.Fatal("second MarkViewed must be a no-op")
	}
	if r.ViewedAt == nil || !r.ViewedAt.Equal(first) {
		t.Fatalf("first receipt time overwritten: got %v want %v", r.ViewedAt, first)
	}
	if r.Version != 2 {
		t.Fatalf("version bumped by repeated view: %d, want 2", r.Version)
	}
}

func TestForRepresentativeDetachesFileIDs(t *testing.T) {
	r, err := NewReply("r", "p", "officer", 1, KindResolved, "done", []string{"f1", "f2"}, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("NewReply: %v", err)
	}
	copy1 := r.ForRepresentative()
	copy2 := r.ForRepresentative()
	// 代表视图的清单与答复内部、以及各次视图之间都不得共享底层切片。
	copy1.FileIDs[0] = "X"
	if r.FileIDs[0] != "f1" {
		t.Fatalf("representative copy leaked into reply: %v", r.FileIDs)
	}
	if copy2.FileIDs[0] != "f1" {
		t.Fatalf("representative copies share buffer: %v", copy2.FileIDs)
	}
}
