// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package email

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/ory/kratos/courier/template"
)

type (
	LoginRegistrationCodeValid struct {
		deps  template.Dependencies
		model *RecoveryCodeValidModel
	}
	LoginRegistrationCodeValidModel struct {
		To           string
		RecoveryCode string
		Identity     map[string]interface{}
	}
)

func NewLoginRegistrationCodeValid(d template.Dependencies, m *RecoveryCodeValidModel) *LoginRegistrationCodeValid {
	return &LoginRegistrationCodeValid{deps: d, model: m}
}

func (t *LoginRegistrationCodeValid) EmailRecipient() (string, error) {
	return t.model.To, nil
}

func (t *LoginRegistrationCodeValid) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(ctx, t.deps, os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)), "login_registration_code/valid/email.subject.gotmpl", "login_registration_code/valid/email.subject*", t.model, t.deps.CourierConfig().CourierTemplatesLoginRegistrationCodeValid(ctx).Subject)

	return strings.TrimSpace(subject), err
}

func (t *LoginRegistrationCodeValid) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.deps, os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)), "login_registration_code/valid/email.body.gotmpl", "login_registration_code/valid/email.body*", t.model, t.deps.CourierConfig().CourierTemplatesLoginRegistrationCodeValid(ctx).Body.HTML)
}

func (t *LoginRegistrationCodeValid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.deps, os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)), "login_registration_code/valid/email.body.plaintext.gotmpl", "login_registration_code/valid/email.body.plaintext*", t.model, t.deps.CourierConfig().CourierTemplatesLoginRegistrationCodeValid(ctx).Body.PlainText)
}

func (t *LoginRegistrationCodeValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.model)
}
