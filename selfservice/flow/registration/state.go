package registration

import "github.com/ory/kratos/selfservice/flow"

// State represents the state of this request:
//
// - choose_method: ask the user to choose a method (e.g. registration with email)
// - sent_email: the email has been sent to the user
// - passed_challenge: the request was successful and the registration challenge was passed.
//
// required: true
type State = flow.State
