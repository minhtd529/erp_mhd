package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
	"github.com/mdh/erp-audit/api/pkg/push"
)

func makePushChallenge(response *string, verified bool, expired bool) *domain.TwoFactorChallenge {
	exp := time.Now().Add(5 * time.Minute)
	if expired {
		exp = time.Now().Add(-1 * time.Minute)
	}
	ch := &domain.TwoFactorChallenge{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		ChallengeID:  uuid.New().String(),
		Method:       "push",
		ExpiresAt:    exp,
		PushResponse: response,
		CreatedAt:    time.Now(),
	}
	if verified {
		now := time.Now()
		ch.VerifiedAt = &now
	}
	return ch
}

func newPush2FAUC(twofa *mockTwoFARepo, users *mockUserRepo) *usecase.Push2FAUseCase {
	return usecase.NewPush2FAUseCase(
		twofa, users, &mockTokenRepo{}, &mockJWTSvc{accessToken: "tok", rawRefresh: "ref"},
		push.NewRelay(),
	)
}

// ── RespondToPush ─────────────────────────────────────────────────────────────

func TestRespondToPush_Happy(t *testing.T) {
	twofa := &mockTwoFARepo{}
	uc := newPush2FAUC(twofa, &mockUserRepo{})
	if err := uc.RespondToPush(context.Background(), "challenge-id", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── GetPushStatus ─────────────────────────────────────────────────────────────

func TestGetPushStatus_Pending(t *testing.T) {
	ch := makePushChallenge(nil, false, false)
	twofa := &mockTwoFARepo{challenge: ch}
	uc := newPush2FAUC(twofa, &mockUserRepo{})

	status, tokens, err := uc.GetPushStatus(context.Background(), ch.ChallengeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.PushChallengePending {
		t.Errorf("expected pending, got %s", status)
	}
	if tokens != nil {
		t.Error("expected no tokens for pending challenge")
	}
}

func TestGetPushStatus_Expired(t *testing.T) {
	ch := makePushChallenge(nil, false, true)
	twofa := &mockTwoFARepo{challenge: ch}
	uc := newPush2FAUC(twofa, &mockUserRepo{})

	status, tokens, err := uc.GetPushStatus(context.Background(), ch.ChallengeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.PushChallengeExpired {
		t.Errorf("expected expired, got %s", status)
	}
	if tokens != nil {
		t.Error("expected no tokens for expired challenge")
	}
}

func TestGetPushStatus_Rejected(t *testing.T) {
	resp := "rejected"
	ch := makePushChallenge(&resp, false, false)
	twofa := &mockTwoFARepo{challenge: ch}
	uc := newPush2FAUC(twofa, &mockUserRepo{})

	status, tokens, err := uc.GetPushStatus(context.Background(), ch.ChallengeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.PushChallengeRejected {
		t.Errorf("expected rejected, got %s", status)
	}
	if tokens != nil {
		t.Error("expected no tokens for rejected challenge")
	}
}

func TestGetPushStatus_ApprovedIssuesTokens(t *testing.T) {
	resp := "approved"
	ch := makePushChallenge(&resp, false, false)
	userID := ch.UserID
	user := &domain.UserForAuth{ID: userID, Email: "sp@test.com", Roles: []string{"AUDIT_STAFF"}}
	twofa := &mockTwoFARepo{challenge: ch}
	userRepo := &mockUserRepo{user: user}
	uc := newPush2FAUC(twofa, userRepo)

	status, tokens, err := uc.GetPushStatus(context.Background(), ch.ChallengeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.PushChallengeApproved {
		t.Errorf("expected approved, got %s", status)
	}
	if tokens == nil {
		t.Fatal("expected tokens on approved challenge")
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestGetPushStatus_ApprovedAlreadyVerified_NoTokens(t *testing.T) {
	resp := "approved"
	ch := makePushChallenge(&resp, true, false) // already verified
	twofa := &mockTwoFARepo{challenge: ch}
	uc := newPush2FAUC(twofa, &mockUserRepo{})

	status, tokens, err := uc.GetPushStatus(context.Background(), ch.ChallengeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.PushChallengeApproved {
		t.Errorf("expected approved, got %s", status)
	}
	if tokens != nil {
		t.Error("already-verified challenge should not re-issue tokens")
	}
}

func TestGetPushStatus_ChallengeNotFound(t *testing.T) {
	twofa := &mockTwoFARepo{findChallengeErr: domain.ErrChallengeNotFound}
	uc := newPush2FAUC(twofa, &mockUserRepo{})

	_, _, err := uc.GetPushStatus(context.Background(), "no-such-id")
	if err == nil {
		t.Fatal("expected error for missing challenge")
	}
}

// ── SendPushToDevice ──────────────────────────────────────────────────────────

func TestSendPushToDevice_Offline(t *testing.T) {
	twofa := &mockTwoFARepo{}
	uc := newPush2FAUC(twofa, &mockUserRepo{})
	// device not connected → should return false, no panic
	delivered := uc.SendPushToDevice("some-token", push.PushPayload{Title: "test"})
	if delivered {
		t.Error("expected offline delivery to return false")
	}
}
