package billing

import "context"

// Provisioner adapts Service onto identity.PersonalProvisioner and org.OrgProvisioner
// (EnsurePersonal/EnsureOrganization returning error only) without import cycles.
type Provisioner struct{ *Service }

func (p Provisioner) EnsurePersonal(ctx context.Context, userID string) error {
	_, _, err := p.Service.EnsurePersonal(ctx, userID)
	return err
}

func (p Provisioner) EnsureOrganization(ctx context.Context, orgID string) error {
	_, _, err := p.Service.EnsureOrganization(ctx, orgID)
	return err
}
