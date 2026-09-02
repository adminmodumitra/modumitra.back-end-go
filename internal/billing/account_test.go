package billing

import (
	"context"
	"testing"
)

func TestEnsurePersonalIdempotent(t *testing.T) {
	svc, ident, mail, _, _, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	a1, s1, err := svc.EnsurePersonal(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if a1.OwnerType != OwnerUser || a1.OwnerID != u.ID || s1.PlanCode != PlanTrial {
		t.Fatalf("acc=%+v sub=%+v", a1, s1)
	}
	a2, s2, err := svc.EnsurePersonal(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if a2.ID != a1.ID || s2.ID != s1.ID {
		t.Fatalf("duplicated personal account %s/%s vs %s/%s", a1.ID, s1.ID, a2.ID, s2.ID)
	}
}

func TestEnsureOrganizationDoesNotTouchPersonal(t *testing.T) {
	svc, ident, mail, orgs, _, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	personal, psub, err := svc.EnsurePersonal(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	o, _, err := orgs.Create(context.Background(), u.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	orgAcc, orgSub, err := svc.EnsureOrganization(context.Background(), o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if orgAcc.ID == personal.ID || orgSub.ID == psub.ID {
		t.Fatalf("org stole personal ba/sub")
	}
	if orgAcc.OwnerType != OwnerOrganization || orgAcc.OwnerID != o.ID {
		t.Fatalf("org acc=%+v", orgAcc)
	}
	again, asub, err := svc.EnsurePersonal(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != personal.ID || asub.ID != psub.ID {
		t.Fatalf("personal changed after org ensure")
	}
}

func TestResolvePersonalVsOrg(t *testing.T) {
	svc, ident, mail, orgs, _, _ := setup(t)
	u := mustUser(t, ident, mail, "ada@example.com")
	o, _, err := orgs.Create(context.Background(), u.ID, "Lovelace Cabinets")
	if err != nil {
		t.Fatal(err)
	}
	personal, err := svc.Resolve(context.Background(), u.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	company, err := svc.Resolve(context.Background(), u.ID, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if personal.ID == company.ID || personal.OwnerType != OwnerUser || company.OwnerType != OwnerOrganization {
		t.Fatalf("personal=%+v company=%+v", personal, company)
	}
}
