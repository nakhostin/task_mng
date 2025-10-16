package entity

type Priority string

const (
	PriorityLowest  Priority = "lowest"
	PriorityLow     Priority = "low"
	PriorityMedium  Priority = "medium"
	PriorityHigh    Priority = "high"
	PriorityHighest Priority = "highest"
)

func (p Priority) String() string {
	return string(p)
}
