// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
)

var _ login.Strategy = new(Strategy)

type loginSubmitPayload struct {
	Method     string `json:"method"`
	CSRFToken  string `json:"csrf_token"`
	Code       string `json:"code"`
	Identifier string `json:"identifier"`
}

func (s *Strategy) RegisterLoginRoutes(*x.RouterPublic) {}

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: identity.CredentialsTypeCodeAuth,
		AAL:    identity.AuthenticatorAssuranceLevel1,
	}
}

func (s *Strategy) HandleLoginError(w http.ResponseWriter, r *http.Request, flow *login.Flow, body *loginSubmitPayload, err error) error {
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

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, lf *login.Flow) error {
	return s.PopulateMethod(r, lf)
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, identityID uuid.UUID) (i *identity.Identity, err error) {
	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, err
	}

	if err := flow.MethodEnabledAndAllowedFromRequest(r, s.ID().String(), s.deps); err != nil {
		return nil, err
	}

	var p loginSubmitPayload
	if err := s.dx.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginMethodSchema),
		decoderx.HTTPDecoderAllowedMethods("POST"),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.HandleLoginError(w, r, f, &p, err)
	}

	if err := flow.EnsureCSRF(s.deps, r, f.Type, s.deps.Config().DisableAPIFlowEnforcement(r.Context()), s.deps.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.HandleLoginError(w, r, f, &p, err)
	}

	i, c, err := s.deps.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), p.Identifier)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			// If return_to was set before, we need to preserve it.
			var opts []registration.FlowOption
			if len(f.ReturnTo) > 0 {
				opts = append(opts, registration.WithFlowReturnTo(f.ReturnTo))
			}

			rf, err := s.deps.RegistrationHandler().NewRegistrationFlow(w, r, f.Type, opts...)
			if err != nil {
				return nil, s.HandleLoginError(w, r, f, &p, err)
			}

			err = s.deps.SessionTokenExchangePersister().MoveToNewFlow(r.Context(), f.ID, rf.ID)
			if err != nil {
				return nil, s.HandleLoginError(w, r, f, &p, err)
			}

			rf.RequestURL, err = x.TakeOverReturnToParameter(f.RequestURL, rf.RequestURL)
			if err != nil {
				return nil, s.HandleLoginError(w, r, f, &p, err)
			}

			// TODO: process registration
		} else {
			return nil, s.HandleLoginError(w, r, f, &p, err)
		}
	}

	var o identity.CredentialsOTP
	d := json.NewDecoder(bytes.NewBuffer(c.Config))
	if err := d.Decode(&o); err != nil {
		return nil, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()).WithWrap(err)
	}

	f.Active = identity.CredentialsTypeCodeAuth
	if err = s.deps.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		return nil, s.HandleLoginError(w, r, f, &p, err)
	}

	return i, nil
}
