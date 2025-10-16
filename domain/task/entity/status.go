package entity

type Status string

const (
	StatusTodo       Status = "ToDo"
	StatusInProgress Status = "InProgress"
	StatusDone       Status = "Done"
)

func (s Status) String() string {
	return string(s)
}
