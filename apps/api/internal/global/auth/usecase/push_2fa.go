package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"github.com/mdh/erp-audit/api/pkg/push"
)

// Push2FAUseCase handles the mobile-side push 2FA flow.
type Push2FAUseCase struct {
	twofa  domain.TwoFARepository
	users  domain.UserRepository
	tokens domain.RefreshTokenRepository
	jwt    domain.JWTIssuer
	relay  *push.Relay
}

func NewPush2FAUseCase(
	twofa domain.TwoFARepository,
	users domain.UserRepository,
	tokens domain.RefreshTokenRepository,
	jwt domain.JWTIssuer,
	relay *push.Relay,
) *Push2FAUseCase {
	return &Push2FAUseCase{twofa: twofa, users: users, tokens: tokens, jwt: jwt, relay: relay}
}

// RespondToPush is called by the mobile app to approve or reject a push 2FA challenge.
func (uc *Push2FAUseCase) RespondToPush(ctx context.Context, challengeID string, approved bool) error {
	return uc.twofa.RespondToPushChallenge(ctx, challengeID, approved)
}

// GetPushStatus is polled by the web browser to check whether the challenge was approved.
// Returns the status and, if just approved, a token pair.
func (uc *Push2FAUseCase) GetPushStatus(ctx context.Context, challengeID string) (domain.PushChallengeStatus, *LoginResponse, error) {
	ch, err := uc.twofa.FindPushChallenge(ctx, challengeID)
	if err != nil {
		return "", nil, err
	}

	if ch.ExpiresAt.Before(time.Now()) {
		return domain.PushChallengeExpired, nil, nil
	}

	if ch.PushResponse == nil {
		return domain.PushChallengePending, nil, nil
	}

	if *ch.PushResponse == "rejected" {
		return domain.PushChallengeRejected, nil, nil
	}

	// Approved — mark verified and issue tokens (idempotent: only if not yet verified)
	if ch.VerifiedAt != nil {
		return domain.PushChallengeApproved, nil, nil
	}

	if err := uc.twofa.MarkChallengeVerified(ctx, challengeID); err != nil {
		return "", nil, err
	}

	user, err := uc.users.FindByID(ctx, ch.UserID)
	if err != nil {
		return "", nil, err
	}

	claims := pkgauth.TokenClaims{
		UserID:       user.ID,
		Email:        user.Email,
		Roles:        user.Roles,
		Permissions:  user.Permissions,
		BranchID:     user.BranchID,
		DepartmentID: user.DepartmentID,
	}
	accessToken, err := uc.jwt.IssueAccessToken(claims)
	if err != nil {
		return "", nil, fmt.Errorf("push_2fa: issue access token: %w", err)
	}

	rawRefresh, _ := uc.jwt.IssueRefreshToken()

	_ = uc.users.UpdateLastLogin(ctx, user.ID)

	return domain.PushChallengeApproved, &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    uc.jwt.AccessTokenTTLSeconds(),
	}, nil
}

// SendPushToDevice delivers a push payload to the device via relay if online.
func (uc *Push2FAUseCase) SendPushToDevice(deviceToken string, payload push.PushPayload) bool {
	return uc.relay.Send(deviceToken, payload)
}
