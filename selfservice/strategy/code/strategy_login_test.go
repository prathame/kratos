package code_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestLoginCodeStrategy(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.StrategyEnable(t, conf, string(identity.CredentialsTypeOTPAuth), true)

	initViper(t, ctx, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	createIdentity := func(t *testing.T) *identity.Identity {
		t.Helper()
		identity := identity.Identity{
			Traits: identity.Traits(fmt.Sprintf(`{"email":"%s"}`, testhelpers.RandomEmail())),
		}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, &identity))
		return &identity
	}

	t.Run("case=should be able to log in with otp without any other identity credentials", func(t *testing.T) {
		identity := createIdentity(t)
		client := testhelpers.NewClientWithCookies(t)

		// 1. Initiate flow
		resp, err := client.Get(public.URL + login.RouteInitBrowserFlow)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		flowID := gjson.GetBytes(body, "id").String()
		require.NotEmpty(t, flowID)

		csrfToken := gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		require.NotEmpty(t, csrfToken)

		require.NoError(t, resp.Body.Close())

		loginEmail := gjson.Get(identity.Traits.String(), "traits.email").String()

		// 2. Submit Identifier (email)
		resp, err = client.PostForm(public.URL+login.RouteSubmitFlow, url.Values{
			"csrf_token": {csrfToken},
			"method":     {"otp"},
			"identifier": {loginEmail},
		})
		require.NoError(t, err)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		csrfToken = gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		require.NotEmpty(t, csrfToken)

		require.NoError(t, resp.Body.Close())

		message := testhelpers.CourierExpectMessage(t, reg, loginEmail, "Login to your account")
		assert.Contains(t, message.Body, "please login to your account by entering the following code")

		loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
		assert.NotEmpty(t, loginCode)

		// 3. Submit OTP
		resp, err = client.PostForm(public.URL+login.RouteSubmitFlow, url.Values{
			"csrf_token": {csrfToken},
			"method":     {"otp"},
			"otp":        {loginCode},
		})
	})
}
