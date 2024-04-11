package config

// MissingConfigInputError is return when a required configuration setting is not found.
type MissingConfigInputError struct {
	env string
}

func (m *MissingConfigInputError) Error() string {
	return "missing env: " + m.env
}

func NewMissingConfigInputError(env string) *MissingConfigInputError {
	return &MissingConfigInputError{env: env}
}

// InvalidConfigInputError is return when a required configuration setting is found but invalid.
type InvalidConfigInputError struct {
	env, reason string
}

func (i *InvalidConfigInputError) Error() string {
	return "invalid env: " + i.env + " reason: " + i.reason
}

func NewInvalidConfigInputError(env, reason string) *InvalidConfigInputError {
	return &InvalidConfigInputError{env: env, reason: reason}
}
