package flow

// Flow State
//
// The state represents the state of the verification flow.
//
// - choose_method: ask the user to choose a method (e.g. recover account via email)
// - sent_email: the email has been sent to the user
// - passed_challenge: the request was successful and the recovery challenge was passed.
type State string

const (
	StateChooseMethod    State = "choose_method"
	StateEmailSent       State = "sent_email"
	StatePassedChallenge State = "passed_challenge"
	StateShowForm        State = "show_form"
	StateSuccess         State = "success"
)

var states = []State{StateChooseMethod, StateEmailSent, StatePassedChallenge}

func indexOf(current State) int {
	for k, s := range states {
		if s == current {
			return k
		}
	}
	return 0
}

func HasReachedState(expected, actual State) bool {
	return indexOf(actual) >= indexOf(expected)
}

func NextState(current State) State {
	if current == StatePassedChallenge {
		return StatePassedChallenge
	}

	return states[indexOf(current)+1]
}
