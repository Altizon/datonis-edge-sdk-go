package edge

type Thing struct {
	Key           string
	DeviceKey     string
	Name          string
	Description   string
	BiDirectional bool
}

func NewThing() *Thing {
	return &Thing{}
}
