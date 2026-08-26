package proposal

import (
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

func TestSubmittedTextStaysFrozenAndSupplementRemainsSeparate(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	p, err := New("p1", "rep1", "老旧小区加装电梯", "出行困难", "建议完善协商机制", "community", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Advance(StatusDraft, StatusSubmitted, now); err != nil {
		t.Fatal(err)
	}
	originalBody := p.Body
	if err := p.Edit(p.Title, p.Cause, "直接改写已提交正文", p.Category, now); !errors.Is(err, shared.ErrInvalidState) {
		t.Fatalf("submitted proposal edit should be rejected, got %v", err)
	}
	supplement := Supplement{ID: "s1", AuthorID: "rep1", Body: "补充居民签字材料", CreatedAt: now}
	if err := p.AddSupplement(supplement, now); err != nil {
		t.Fatal(err)
	}
	if p.Body != originalBody || len(p.Supplements) != 1 || p.Supplements[0].Body != supplement.Body {
		t.Fatalf("supplement replaced frozen text: body=%q supplements=%+v", p.Body, p.Supplements)
	}
	policy := DefaultFilePolicy()
	if err := policy.Validate("evidence.pdf", "application/pdf", 10, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "../private/evidence.pdf"); err == nil {
		t.Fatal("traversal storage key should be rejected")
	}
}
