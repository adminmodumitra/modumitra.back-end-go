package outbox

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Danish-13-Khan/backend-cabinet/internal/obs"
	"github.com/Danish-13-Khan/backend-cabinet/internal/outbox/pb"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func meta(ctx context.Context, eventType, partitionKey string, at time.Time) *pb.Envelope {
	return &pb.Envelope{
		EventType:    eventType,
		PartitionKey: partitionKey,
		TraceId:      obs.TraceIDFromContext(ctx),
		RequestId:    utils.RequestIDFromContext(ctx),
		OccurredAt:   timestamppb.New(at.UTC()),
	}
}

func UserCreated(ctx context.Context, userID string, at time.Time) *pb.Envelope {
	env := meta(ctx, EventUserCreated, userID, at)
	env.Body = &pb.Envelope_UserCreated{UserCreated: &pb.UserCreated{UserId: userID}}
	return env
}

func AuthLogin(ctx context.Context, userID, sessionID string, at time.Time) *pb.Envelope {
	env := meta(ctx, EventAuthLogin, userID, at)
	env.Body = &pb.Envelope_AuthLogin{AuthLogin: &pb.AuthLogin{UserId: userID, SessionId: sessionID}}
	return env
}

func OrgCreated(ctx context.Context, orgID, ownerUserID, name string, at time.Time) *pb.Envelope {
	env := meta(ctx, EventOrgCreated, orgID, at)
	env.Body = &pb.Envelope_OrgCreated{OrgCreated: &pb.OrgCreated{
		OrganizationId: orgID,
		OwnerUserId:    ownerUserID,
		Name:           name,
	}}
	return env
}

func MembershipCreated(ctx context.Context, membershipID, orgID, userID, role string, at time.Time) *pb.Envelope {
	env := meta(ctx, EventMembershipCreated, orgID, at)
	env.Body = &pb.Envelope_MembershipCreated{MembershipCreated: &pb.MembershipCreated{
		MembershipId:   membershipID,
		OrganizationId: orgID,
		UserId:         userID,
		Role:           role,
	}}
	return env
}

func BillingTrialStarted(ctx context.Context, orgID, ownerUserID string, trialEndsAt, at time.Time) *pb.Envelope {
	env := meta(ctx, EventBillingTrialStarted, orgID, at)
	env.Body = &pb.Envelope_BillingTrialStarted{BillingTrialStarted: &pb.BillingTrialStarted{
		OrganizationId: orgID,
		OwnerUserId:    ownerUserID,
		TrialEndsAt:    timestamppb.New(trialEndsAt.UTC()),
	}}
	return env
}

func BillingSubscriptionUpdated(ctx context.Context, orgID, subID, planID, status string, at time.Time) *pb.Envelope {
	env := meta(ctx, EventBillingSubscriptionUpdated, orgID, at)
	env.Body = &pb.Envelope_BillingSubscriptionUpdated{BillingSubscriptionUpdated: &pb.BillingSubscriptionUpdated{
		OrganizationId: orgID,
		SubscriptionId: subID,
		PlanId:         planID,
		Status:         status,
	}}
	return env
}

func DeviceActivated(ctx context.Context, deviceID, userID, platform string, at time.Time) *pb.Envelope {
	env := meta(ctx, EventDeviceActivated, userID, at)
	env.Body = &pb.Envelope_DeviceActivated{DeviceActivated: &pb.DeviceActivated{
		DeviceId: deviceID,
		UserId:   userID,
		Platform: platform,
	}}
	return env
}

func InvitationAccepted(ctx context.Context, invitationID, orgID, userID, role string, at time.Time) *pb.Envelope {
	env := meta(ctx, EventInvitationAccepted, orgID, at)
	env.Body = &pb.Envelope_InvitationAccepted{InvitationAccepted: &pb.InvitationAccepted{
		InvitationId:   invitationID,
		OrganizationId: orgID,
		UserId:         userID,
		Role:           role,
	}}
	return env
}

// Enqueue is a nil-safe helper so services can skip when no writer is injected.
func Enqueue(ctx context.Context, w Writer, env *pb.Envelope) error {
	if w == nil || env == nil {
		return nil
	}
	return w.Enqueue(ctx, env)
}
