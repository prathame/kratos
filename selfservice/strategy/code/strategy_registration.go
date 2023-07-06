// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"encoding/json"
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

var _ registration.Strategy = new(Strategy)

// Update Registration Flow with Code Method
//
// swagger:model updateRegistrationFlowWithCodeMethod
type UpdateRegistrationFlowWithCodeMethod struct {
	// Password to sign the user up with
	//
	// required: true
	Password string `json:"password"`

	// The identity's traits
	//
	// required: true
	Traits json.RawMessage `json:"traits"`

	// The CSRF Token
	CSRFToken string `json:"csrf_token"`

	// Method to use
	//
	// This field must be set to `code` when using the code method.
	//
	// required: true
	Method string `json:"method"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty"`
}

func (s *Strategy) RegisterRegistrationRoutes(*x.RouterPublic) {}

func (s *Strategy) HandleRegistrationError(w http.ResponseWriter, r *http.Request, flow *registration.Flow, body *UpdateRegistrationFlowWithCodeMethod, err error) error {
	if flow != nil {
		email := ""
		if body != nil {
			email = body.Traits
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

	var p UpdateRegistrationFlowWithCodeMethod
	if err := registration.DecodeBody(&p, r, s.dx, s.deps.Config(), registrationSchema); err != nil {
		return s.HandleRegistrationError(w, r, f, &p, err)
	}

	f.TransientPayload = p.TransientPayload

	if err := flow.EnsureCSRF(s.deps, r, f.Type, s.deps.Config().DisableAPIFlowEnforcement(r.Context()), s.deps.GenerateCSRFToken, p.CSRFToken); err != nil {
		return s.HandleRegistrationError(w, r, f, &p, err)
	}

	if len(p.Traits) == 0 {
		p.Traits = json.RawMessage("{}")
	}

	i.Traits = identity.Traits(p.Traits)
	if err := i.SetCredentialsWithConfig(s.ID(), identity.Credentials{Type: s.ID(), Identifiers: []string{}}, &identity.CredentialsCode{CodeHMAC: ""}); err != nil {
		return s.HandleRegistrationError(w, r, f, &p, err)
	}

	return nil
}
