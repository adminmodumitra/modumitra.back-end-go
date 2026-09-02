package httpapi

import (
	"github.com/Danish-13-Khan/backend-cabinet/internal/models/dto"
	"github.com/Danish-13-Khan/backend-cabinet/internal/sso"
)

func identityView(id *sso.Identity) dto.IdentityView {
	if id == nil {
		return dto.IdentityView{}
	}
	var email *string
	if id.Email != "" {
		e := id.Email
		email = &e
	}
	return dto.IdentityView{
		Provider:       id.Provider,
		ProviderUserID: id.ProviderUserID,
		Email:          email,
		LinkedAt:       rfcTime(id.LinkedAt),
	}
}
