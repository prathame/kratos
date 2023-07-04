// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

// CredentialsOTP represents an OTP code
//
// swagger:model identityCredentialsOTP
type CredentialsOTP struct {
	// CodeHMAC represents the HMACed value of the login/registration code
	CodeHMAC string `json:"code_hmac"`
}
