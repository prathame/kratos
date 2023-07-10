package code

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
)

type RegistrationCodes struct {
	// ID represents the tokens's unique ID.
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// CodeHMAC represents the HMACed value of the verification code
	CodeHMAC string `json:"-" db:"code_hmac"`

	// UsedAt is the timestamp of when the code was used or null if it wasn't yet
	UsedAt sql.NullTime `json:"-" db:"used_at"`

	// ExpiresAt is the time (UTC) when the token expires.
	// required: true
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the token was issued.
	// required: true
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID     uuid.NullUUID `json:"-" faker:"-" db:"selfservice_registration_flow_id"`
	NID        uuid.UUID     `json:"-"  faker:"-" db:"nid"`
	IdentityID uuid.UUID     `json:"identity_id"  faker:"-" db:"identity_id"`
}

func (RegistrationCodes) TableName(ctx context.Context) string {
	return "identity_registration_codes"
}
