package identity

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func testSvc(t *testing.T) (*Service, *RecordingMailer) {
	t.Helper()
	cfg := Settings{
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
		AccessTTL:      15 * time.Minute,
		RefreshTTL:     30 * 24 * time.Hour,
		EmailTokenTTL:  24 * time.Hour,
		JWTSecret:      "test-secret-for-cabinet-studio-hs256",
		Issuer:         Issuer,
	}
	svc, mail, _ := NewMemoryService(cfg)
	return svc, mail
}

func mustRegister(t *testing.T, svc *Service, email, password string) *User {
	t.Helper()
	u, err := svc.Register(context.Background(), RegisterInput{
		Email:          email,
		Password:       password,
		DisplayName:    "Ada Lovelace",
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return u
}

func codeOf(err error) utils.Code {
	var e *utils.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

func TestRegisterSuccess(t *testing.T) {
	svc, mail := testSvc(t)
	u := mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	if u.ID == "" || !strings.HasPrefix(u.ID, "usr_") {
		t.Fatalf("id=%q", u.ID)
	}
	if u.EmailVerified() {
		t.Fatal("should be unverified")
	}
	if u.PasswordHash != "" {
		t.Fatal("password hash leaked on user")
	}
	if _, ok := mail.LastVerification(); !ok {
		t.Fatal("expected verification mail")
	}
}

func TestRegisterEmailTaken(t *testing.T) {
	svc, _ := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	_, err := svc.Register(context.Background(), RegisterInput{
		Email:          "ADA@example.com",
		Password:       "another-horse-battery-19",
		DisplayName:    "Ada",
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
	})
	if codeOf(err) != utils.CodeEmailTaken {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	svc, _ := testSvc(t)
	_, err := svc.Register(context.Background(), RegisterInput{
		Email:          "ada@example.com",
		Password:       "short",
		DisplayName:    "Ada",
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
	})
	if codeOf(err) != utils.CodeWeakPassword {
		t.Fatalf("err=%v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	svc, _ := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	u, pair, err := svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "correct-horse-battery-staple-19",
	})
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken == "" || !strings.HasPrefix(pair.RefreshToken, "rt_") {
		t.Fatalf("pair=%+v", pair)
	}
	if pair.ExpiresIn != 900 || pair.TokenType != "Bearer" {
		t.Fatalf("pair=%+v", pair)
	}
	if u.Email != "ada@example.com" {
		t.Fatalf("email=%s", u.Email)
	}
}

func TestLoginInvalidCredentialsUnknownAndBadPassword(t *testing.T) {
	svc, _ := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")

	_, _, errUnknown := svc.Login(context.Background(), LoginInput{
		Email:    "missing@example.com",
		Password: "correct-horse-battery-staple-19",
	})
	_, _, errBad := svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "wrong-password-xx",
	})
	if codeOf(errUnknown) != utils.CodeInvalidCredentials {
		t.Fatalf("unknown=%v", errUnknown)
	}
	if codeOf(errBad) != utils.CodeInvalidCredentials {
		t.Fatalf("bad=%v", errBad)
	}
	if errUnknown.Error() != errBad.Error() {
		t.Fatalf("messages differ: %q vs %q", errUnknown, errBad)
	}
}

func TestRefreshRotationAndReuse(t *testing.T) {
	svc, _ := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	_, pair, err := svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "correct-horse-battery-staple-19",
	})
	if err != nil {
		t.Fatal(err)
	}
	old := pair.RefreshToken
	next, err := svc.Refresh(context.Background(), old)
	if err != nil {
		t.Fatal(err)
	}
	if next.RefreshToken == "" || next.RefreshToken == old {
		t.Fatalf("did not rotate: %q", next.RefreshToken)
	}
	_, err = svc.Refresh(context.Background(), old)
	if codeOf(err) != utils.CodeTokenReuse {
		t.Fatalf("reuse=%v code=%s", err, codeOf(err))
	}
	// Family revoked: the rotated current token is no longer usable.
	_, err = svc.Refresh(context.Background(), next.RefreshToken)
	if codeOf(err) != utils.CodeUnauthenticated && codeOf(err) != utils.CodeTokenReuse {
		t.Fatalf("family not revoked: %v code=%s", err, codeOf(err))
	}
}

func TestLogout(t *testing.T) {
	svc, _ := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	_, pair, err := svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "correct-horse-battery-staple-19",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(context.Background(), "rt_unknown"); err != nil {
		t.Fatalf("unknown logout: %v", err)
	}
	_, err = svc.Refresh(context.Background(), pair.RefreshToken)
	if codeOf(err) != utils.CodeTokenReuse && codeOf(err) != utils.CodeUnauthenticated {
		t.Fatalf("refresh after logout: %v code=%s", err, codeOf(err))
	}
}

func TestVerifyEmail(t *testing.T) {
	svc, mail := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	call, ok := mail.LastVerification()
	if !ok || !strings.HasPrefix(call.Token, "ev_") {
		t.Fatalf("mail=%+v ok=%v", call, ok)
	}
	u, err := svc.VerifyEmail(context.Background(), call.Token)
	if err != nil {
		t.Fatal(err)
	}
	if !u.EmailVerified() {
		t.Fatal("not verified")
	}
	_, err = svc.VerifyEmail(context.Background(), call.Token)
	if codeOf(err) != utils.CodeInvalidToken {
		t.Fatalf("reuse token: %v", err)
	}
}

func TestPasswordResetConfirmRevokesSessions(t *testing.T) {
	svc, mail := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	_, pair, err := svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "correct-horse-battery-staple-19",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.RequestPasswordReset(context.Background(), "ada@example.com"); err != nil {
		t.Fatal(err)
	}
	call, ok := mail.LastPasswordReset()
	if !ok || !strings.HasPrefix(call.Token, "pr_") {
		t.Fatalf("reset mail=%+v", call)
	}
	if err := svc.ConfirmPasswordReset(context.Background(), call.Token, "new-correct-horse-battery-21"); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Refresh(context.Background(), pair.RefreshToken)
	if codeOf(err) != utils.CodeTokenReuse && codeOf(err) != utils.CodeUnauthenticated {
		t.Fatalf("session still live: %v code=%s", err, codeOf(err))
	}
	_, _, err = svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "new-correct-horse-battery-21",
	})
	if err != nil {
		t.Fatalf("login with new password: %v", err)
	}
}

func TestChangePassword(t *testing.T) {
	svc, _ := testSvc(t)
	u := mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	_, pair, err := svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "correct-horse-battery-staple-19",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = svc.ChangePassword(context.Background(), u.ID, "wrong-current-xx", "new-correct-horse-battery-21")
	if codeOf(err) != utils.CodeInvalidCredentials {
		t.Fatalf("wrong current: %v", err)
	}
	if err := svc.ChangePassword(context.Background(), u.ID, "correct-horse-battery-staple-19", "new-correct-horse-battery-21"); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Refresh(context.Background(), pair.RefreshToken)
	if codeOf(err) != utils.CodeTokenReuse && codeOf(err) != utils.CodeUnauthenticated {
		t.Fatalf("sessions not revoked: %v", err)
	}
	_, _, err = svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "new-correct-horse-battery-21",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMe(t *testing.T) {
	svc, _ := testSvc(t)
	u := mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	got, err := svc.Me(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != "ada@example.com" || got.PasswordHash != "" {
		t.Fatalf("%+v", got)
	}
}

func TestRegisterTermsOutdated(t *testing.T) {
	svc, _ := testSvc(t)
	_, err := svc.Register(context.Background(), RegisterInput{
		Email:          "ada@example.com",
		Password:       "correct-horse-battery-staple-19",
		DisplayName:    "Ada",
		TermsVersion:   "1999-01-01",
		PrivacyVersion: "2026-06-01",
	})
	if codeOf(err) != utils.CodeTermsOutdated {
		t.Fatalf("err=%v", err)
	}
}

func TestLoginSuspended(t *testing.T) {
	svc, _, mem := NewMemoryService(Settings{
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
		AccessTTL:      15 * time.Minute,
		RefreshTTL:     30 * 24 * time.Hour,
		JWTSecret:      "test-secret-for-cabinet-studio-hs256",
	})
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	mem.mu.Lock()
	for _, u := range mem.users {
		u.Status = StatusDisabled
	}
	mem.mu.Unlock()
	_, _, err := svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "correct-horse-battery-staple-19",
	})
	if codeOf(err) != utils.CodeAccountSuspended {
		t.Fatalf("err=%v", err)
	}
}

func TestRequestOTPUnknownEmail(t *testing.T) {
	svc, mail := testSvc(t)
	if err := svc.RequestOTP(context.Background(), "nobody@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, ok := mail.LastOTP(); ok {
		t.Fatal("expected no otp mail for unknown email")
	}
}

func TestRequestOTPKnownEmailRecordsSixDigit(t *testing.T) {
	svc, mail := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	if err := svc.RequestOTP(context.Background(), "ada@example.com"); err != nil {
		t.Fatal(err)
	}
	call, ok := mail.LastOTP()
	if !ok {
		t.Fatal("expected otp mail")
	}
	if call.Email != "ada@example.com" {
		t.Fatalf("email=%s", call.Email)
	}
	if !validOTP(call.Token) {
		t.Fatalf("otp=%q", call.Token)
	}
}

func TestVerifyOTPSuccessIssuesTokens(t *testing.T) {
	svc, mail := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	if err := svc.RequestOTP(context.Background(), "ada@example.com"); err != nil {
		t.Fatal(err)
	}
	call, ok := mail.LastOTP()
	if !ok {
		t.Fatal("expected otp mail")
	}
	u, pair, err := svc.VerifyOTP(context.Background(), "ada@example.com", call.Token, "Ada’s Studio Mac")
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken == "" || !strings.HasPrefix(pair.RefreshToken, "rt_") {
		t.Fatalf("pair=%+v", pair)
	}
	if pair.ExpiresIn != 900 || pair.TokenType != "Bearer" {
		t.Fatalf("pair=%+v", pair)
	}
	if u.Email != "ada@example.com" || u.PasswordHash != "" {
		t.Fatalf("user=%+v", u)
	}
	if u.EmailVerified() {
		t.Fatal("unverified users may still get tokens")
	}
}

func TestVerifyOTPWrongCode(t *testing.T) {
	svc, mail := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	if err := svc.RequestOTP(context.Background(), "ada@example.com"); err != nil {
		t.Fatal(err)
	}
	call, _ := mail.LastOTP()
	bad := "000000"
	if call.Token == bad {
		bad = "000001"
	}
	_, _, err := svc.VerifyOTP(context.Background(), "ada@example.com", bad, "")
	if codeOf(err) != utils.CodeInvalidCredentials {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestVerifyOTPReuse(t *testing.T) {
	svc, mail := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	if err := svc.RequestOTP(context.Background(), "ada@example.com"); err != nil {
		t.Fatal(err)
	}
	call, ok := mail.LastOTP()
	if !ok {
		t.Fatal("expected otp mail")
	}
	if _, _, err := svc.VerifyOTP(context.Background(), "ada@example.com", call.Token, ""); err != nil {
		t.Fatal(err)
	}
	_, _, err := svc.VerifyOTP(context.Background(), "ada@example.com", call.Token, "")
	if codeOf(err) != utils.CodeInvalidCredentials {
		t.Fatalf("reuse=%v code=%s", err, codeOf(err))
	}
}

func TestRequestOTPResendInvalidatesPrevious(t *testing.T) {
	svc, mail := testSvc(t)
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	if err := svc.RequestOTP(context.Background(), "ada@example.com"); err != nil {
		t.Fatal(err)
	}
	first, ok := mail.LastOTP()
	if !ok {
		t.Fatal("expected first otp")
	}
	var second MailCall
	for i := 0; i < 8; i++ {
		if err := svc.RequestOTP(context.Background(), "ada@example.com"); err != nil {
			t.Fatal(err)
		}
		call, ok := mail.LastOTP()
		if !ok {
			t.Fatal("expected second otp")
		}
		second = call
		if second.Token != first.Token {
			break
		}
	}
	if first.Token == second.Token {
		t.Fatal("expected a new otp on resend")
	}
	_, _, err := svc.VerifyOTP(context.Background(), "ada@example.com", first.Token, "")
	if codeOf(err) != utils.CodeInvalidCredentials {
		t.Fatalf("old otp=%v code=%s", err, codeOf(err))
	}
	if _, _, err := svc.VerifyOTP(context.Background(), "ada@example.com", second.Token, ""); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyOTPUnknownEmail(t *testing.T) {
	svc, _ := testSvc(t)
	_, _, err := svc.VerifyOTP(context.Background(), "nobody@example.com", "123456", "")
	if codeOf(err) != utils.CodeInvalidCredentials {
		t.Fatalf("err=%v code=%s", err, codeOf(err))
	}
}

func TestVerifyOTPFormatAndEmptyEmail(t *testing.T) {
	svc, _ := testSvc(t)
	_, _, err := svc.VerifyOTP(context.Background(), "", "123456", "")
	if codeOf(err) != utils.CodeValidation {
		t.Fatalf("empty email=%v", err)
	}
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	_, _, err = svc.VerifyOTP(context.Background(), "ada@example.com", "12345", "")
	if codeOf(err) != utils.CodeValidation {
		t.Fatalf("short otp=%v", err)
	}
	_, _, err = svc.VerifyOTP(context.Background(), "ada@example.com", "12345a", "")
	if codeOf(err) != utils.CodeValidation {
		t.Fatalf("non-digit otp=%v", err)
	}
}

func TestVerifyOTPSuspended(t *testing.T) {
	svc, mail, mem := NewMemoryService(Settings{
		TermsVersion:   "2026-06-01",
		PrivacyVersion: "2026-06-01",
		AccessTTL:      15 * time.Minute,
		RefreshTTL:     30 * 24 * time.Hour,
		JWTSecret:      "test-secret-for-cabinet-studio-hs256",
	})
	mustRegister(t, svc, "ada@example.com", "correct-horse-battery-staple-19")
	if err := svc.RequestOTP(context.Background(), "ada@example.com"); err != nil {
		t.Fatal(err)
	}
	call, ok := mail.LastOTP()
	if !ok {
		t.Fatal("expected otp")
	}
	mem.mu.Lock()
	for _, u := range mem.users {
		u.Status = StatusDisabled
	}
	mem.mu.Unlock()
	_, _, err := svc.VerifyOTP(context.Background(), "ada@example.com", call.Token, "")
	if codeOf(err) != utils.CodeAccountSuspended {
		t.Fatalf("err=%v", err)
	}
}
