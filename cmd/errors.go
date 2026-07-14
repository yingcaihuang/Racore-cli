package cmd

// AuthError represents an authentication failure (exit code 1).
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// NetworkError represents a network or timeout failure (exit code 2).
type NetworkError struct {
	Message string
}

func (e *NetworkError) Error() string {
	return e.Message
}

// InputError represents invalid input or missing arguments (exit code 3).
type InputError struct {
	Message string
}

func (e *InputError) Error() string {
	return e.Message
}
