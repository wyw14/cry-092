package response

import (
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Kind string

const (
	KindResolved      Kind = "resolved"
	KindPlanned       Kind = "planned"
	KindExplained     Kind = "explained"
	KindReferenceOnly Kind = "reference_only"
)

type Reply struct {
	ID          string
	ProposalID  string
	Round       int
	Kind        Kind
	Summary     string
	FileIDs     []string
	SubmittedBy string
	SubmittedAt time.Time
	ViewedAt    *time.Time
	Supplements []string
	Version     int64
}

func NewReply(id, proposalID, actorID string, round int, kind Kind, summary string, files []string, now time.Time) (*Reply, error) {
	if round < 1 || summary == "" || len(files) == 0 {
		return nil, shared.NewError("REPLY_INVALID", "reply kind, summary and file are required", nil)
	}
	// 复制成答复自有的切片，断开与调用方入参的底层共享，确保答复证据在提交后不可被外部改动替换。
	ids := make([]string, len(files))
	copy(ids, files)
	return &Reply{ID: id, ProposalID: proposalID, Round: round, Kind: kind, Summary: summary, FileIDs: ids, SubmittedBy: actorID, SubmittedAt: now.UTC(), Version: 1}, nil
}

func validKind(k Kind) bool {
	return k == KindResolved || k == KindPlanned || k == KindExplained || k == KindReferenceOnly
}

func (r *Reply) MarkViewed(now time.Time) bool {
	if r.ViewedAt != nil {
		// 首次查收记录一经写入即不可变，代表再次打开页面不得覆盖该时间。
		return false
	}
	at := now.UTC()
	r.ViewedAt = &at
	r.Version++
	return true
}

func (r *Reply) AddSupplement(text string) error {
	if text == "" {
		return shared.NewError("REPLY_SUPPLEMENT_EMPTY", "supplement cannot be empty", nil)
	}
	r.Supplements = []string{text}
	r.Version++
	return nil
}

type RepresentativeCopy struct {
	ReplyID     string
	Status      string
	Kind        Kind
	Summary     string
	FileIDs     []string
	ViewedAt    *time.Time
	Supplements []string
}

func (r Reply) ForRepresentative() RepresentativeCopy {
	status := "unread"
	if r.ViewedAt != nil {
		status = "viewed"
	}
	// 正式答复文件清单是答复证据，向代表视图下发时同样要复制独立切片，避免调用方再改动代表看到的清单。
	return RepresentativeCopy{
		ReplyID:     r.ID,
		Status:      status,
		Kind:        r.Kind,
		Summary:     r.Summary,
		FileIDs:     copyStrings(r.FileIDs),
		ViewedAt:    r.ViewedAt,
		Supplements: copyStrings(r.Supplements),
	}
}

func copyStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
