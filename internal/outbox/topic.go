package outbox

import "strings"

const (
	TopicOrg     = "cabinet.org.v1"
	TopicAuth    = "cabinet.auth.v1"
	TopicBilling = "cabinet.billing.v1"
	TopicDevice  = "cabinet.device.v1"

	EventUserCreated                = "user.created"
	EventAuthLogin                  = "auth.login"
	EventOrgCreated                 = "org.created"
	EventMembershipCreated          = "membership.created"
	EventBillingTrialStarted        = "billing.trial_started"
	EventBillingSubscriptionUpdated = "billing.subscription_updated"
	EventDeviceActivated            = "device.activated"
	EventInvitationAccepted         = "invitation.accepted"

	HeaderContentType = "content-type"
	HeaderEventID     = "event_id"
	HeaderEventType   = "event_type"
	HeaderTraceparent = "traceparent"

	ContentTypeProtobuf = "application/protobuf"
)

// TopicFor maps event_type to a Kafka topic. Prefix is always cabinet.
func TopicFor(eventType string) string {
	eventType = strings.TrimSpace(eventType)
	switch {
	case strings.HasPrefix(eventType, "org."),
		strings.HasPrefix(eventType, "membership."),
		strings.HasPrefix(eventType, "invitation."):
		return TopicOrg
	case strings.HasPrefix(eventType, "auth."),
		strings.HasPrefix(eventType, "user."):
		return TopicAuth
	case strings.HasPrefix(eventType, "billing."):
		return TopicBilling
	case strings.HasPrefix(eventType, "device."):
		return TopicDevice
	default:
		head, _, _ := strings.Cut(eventType, ".")
		if head == "" {
			head = "events"
		}
		return "cabinet." + head + ".v1"
	}
}
