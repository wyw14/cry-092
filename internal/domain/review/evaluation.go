package review

import (
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Rating string

const (
	RatingSatisfied       Rating = "satisfied"
	RatingMostlySatisfied Rating = "mostly_satisfied"
	RatingUnsatisfied     Rating = "unsatisfied"
)

type Evaluation struct {
	ID               string
	ProposalID       string
	ReplyID          string
	Round            int
	RepresentativeID string
	Rating           Rating
	Comment          string
	CreatedAt        time.Time
}

func NewEvaluation(id, proposalID, replyID, representativeID string, round int, rating Rating, comment string, now time.Time) (*Evaluation, error) {
	if round < 1 || (rating != RatingSatisfied && rating != RatingMostlySatisfied && rating != RatingUnsatisfied) {
		return nil, shared.NewError("EVALUATION_INVALID", "evaluation rating is invalid", nil)
	}
	if rating == RatingUnsatisfied && comment == "" {
		return nil, shared.NewError("EVALUATION_REASON_REQUIRED", "unsatisfied evaluation requires a reason", nil)
	}
	return &Evaluation{ID: id, ProposalID: proposalID, ReplyID: replyID, Round: round, RepresentativeID: representativeID, Rating: rating, Comment: comment, CreatedAt: now.UTC()}, nil
}
