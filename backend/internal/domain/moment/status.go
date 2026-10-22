package domain

type Status string

const (
	StatusNone    Status = ""
	StatusInClass Status = "in_class"
	StatusMissYou Status = "miss_you"
	StatusTired   Status = "tired"
)

func (s Status) Valid() bool {
	return s == StatusNone || s == StatusInClass || s == StatusMissYou || s == StatusTired
}
