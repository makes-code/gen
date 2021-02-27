package types

type Prebuilder interface {
	Prebuild() error
}

func prebuild(builder interface{}) error {
	if b, ok := builder.(Prebuilder); ok {
		return b.Prebuild()
	}
	return nil
}
