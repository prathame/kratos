package code

import (
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

var _ registration.Strategy = new(Strategy)

func (s *Strategy) RegisterRegistrationRoutes(*x.RouterPublic) {}

func (s *Strategy) HandleRegistrationError(w http.ResponseWriter, r *http.Request, flow *registration.Flow, body *loginSubmitPayload, err error) error {
	if flow != nil {
		email := ""
		if body != nil {
			email = body.Identifier
		}

		flow.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
		flow.UI.GetNodes().Upsert(
			node.NewInputField("identifier", email, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewInfoNodeInputEmail()),
		)
	}

	return err
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, rf *registration.Flow) error {
	return s.PopulateMethod(r, rf)
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	if err := flow.MethodEnabledAndAllowedFromRequest(r, s.ID().String(), s.deps); err != nil {
		return err
	}

	return nil
}
