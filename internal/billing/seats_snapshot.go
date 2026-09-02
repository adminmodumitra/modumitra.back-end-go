package billing

import (
	"context"

	"github.com/Danish-13-Khan/backend-cabinet/internal/org"
)

func (s *Service) seatSnapshot(ctx context.Context, orgID string, purchased int) (*Seats, error) {
	members, _, err := s.members.ListByOrg(ctx, orgID, "", "", "", 500)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	pending := 0
	if s.invites != nil {
		invs, _, err := s.invites.ListByOrg(ctx, orgID, org.InvitePending, "", 500)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		pending = len(invs)
	}
	out := &Seats{Purchased: purchased, PendingInvitations: pending, Members: []SeatMember{}}
	for i := range members {
		mem := members[i]
		if mem.Status == org.MemberRemoved || mem.Status == org.MemberInvited {
			continue
		}
		consumes := mem.Status == org.MemberActive
		if consumes {
			out.Used++
		}
		out.Members = append(out.Members, SeatMember{
			UserID: mem.UserID, Role: mem.Role, Status: mem.Status, ConsumesSeat: consumes,
		})
	}
	out.Available = purchased - out.Used
	return out, nil
}

func personalSeats(userID string, purchased int) *Seats {
	used := 1
	avail := purchased - used
	return &Seats{
		Purchased: purchased, Used: used, PendingInvitations: 0, Available: avail,
		Members: []SeatMember{{UserID: userID, Role: "owner", Status: "active", ConsumesSeat: true}},
	}
}
