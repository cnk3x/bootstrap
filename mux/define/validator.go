package define

type Validator interface {
	Exec(value string) error
}

type VFunc func(value string) error

func (v VFunc) Exec(value string) error {
	return v(value)
}
