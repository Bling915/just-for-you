package domain

type Photo struct {
	objectKey string
	order     int
}

func NewPhoto(objectKey string, order int) Photo { return Photo{objectKey: objectKey, order: order} }
func (p Photo) ObjectKey() string                { return p.objectKey }
func (p Photo) Order() int                       { return p.order }
