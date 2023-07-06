// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"net/http"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/randx"
)

var (
	_ recovery.Strategy      = new(Strategy)
	_ recovery.AdminHandler  = new(Strategy)
	_ recovery.PublicHandler = new(Strategy)
)

var (
	_ verification.Strategy      = new(Strategy)
	_ verification.AdminHandler  = new(Strategy)
	_ verification.PublicHandler = new(Strategy)
)

var (
	_ login.Strategy        = new(Strategy)
	_ registration.Strategy = new(Strategy)
)

type (
	// FlowMethod contains the configuration for this selfservice strategy.
	FlowMethod struct {
		*container.Container
	}

	strategyDependencies interface {
		x.CSRFProvider
		x.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.LoggingProvider

		config.Provider

		session.HandlerProvider
		session.ManagementProvider
		settings.HandlerProvider
		settings.FlowPersistenceProvider

		identity.ValidationProvider
		identity.ManagementProvider
		identity.PoolProvider
		identity.PrivilegedPoolProvider

		courier.Provider

		errorx.ManagementProvider

		recovery.ErrorHandlerProvider
		recovery.FlowPersistenceProvider
		recovery.StrategyProvider
		recovery.HookExecutorProvider

		verification.FlowPersistenceProvider
		verification.StrategyProvider
		verification.HookExecutorProvider

		login.StrategyProvider
		login.HookExecutorProvider
		login.FlowPersistenceProvider

		registration.StrategyProvider
		registration.HookExecutorProvider
		registration.HandlerProvider

		RecoveryCodePersistenceProvider
		VerificationCodePersistenceProvider
		SenderProvider

		schema.IdentityTraitsProvider

		sessiontokenexchange.PersistenceProvider
	}

	Strategy struct {
		deps strategyDependencies
		dx   *decoderx.HTTP
	}
)

func NewStrategy(deps strategyDependencies) *Strategy {
	return &Strategy{deps: deps, dx: decoderx.NewHTTP()}
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeCodeAuth
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.CodeGroup
}

func (s *Strategy) PopulateMethod(r *http.Request, f flow.Flow) error {
	switch f.GetState() {
	case flow.StateChooseMethod:

		if f.GetFlowName() == flow.VerificationFlow || f.GetFlowName() == flow.RecoveryFlow || f.GetFlowName() == flow.RegistrationFlow {
			f.GetUI().GetNodes().Upsert(
				node.NewInputField("email", nil, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
					WithMetaLabel(text.NewInfoNodeInputEmail()),
			)
		} else if f.GetFlowName() == flow.LoginFlow {
			f.GetUI().GetNodes().Upsert(
				node.NewInputField("identifier", nil, node.CodeGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).
					WithMetaLabel(text.NewInfoNodeInputEmail()),
			)
		}

		break
	case flow.StateEmailSent:
		f.GetUI().Nodes.Upsert(
			node.
				NewInputField("code", nil, node.CodeGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewInfoNodeLabelVerifyOTP()),
		)
		// Required for the re-send code button
		f.GetUI().Nodes.Append(
			node.NewInputField("method", s.NodeGroup(), node.CodeGroup, node.InputAttributeTypeHidden),
		)

		if f.GetFlowName() == flow.VerificationFlow {
			f.GetUI().Messages.Set(text.NewVerificationEmailWithCodeSent())
		}
		break
	}

	f.GetUI().Nodes.Append(
		node.NewInputField("method", s.VerificationStrategyID(), node.CodeGroup, node.InputAttributeTypeSubmit).
			WithMetaLabel(text.NewInfoNodeLabelSubmit()),
	)

	if f.GetType() == flow.TypeBrowser {
		f.GetUI().SetCSRF(s.deps.GenerateCSRFToken(r))
	}
	return nil
}

const CodeLength = 6

func GenerateCode() string {
	return randx.MustString(CodeLength, randx.Numeric)
}
