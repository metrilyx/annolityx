package annotations

const (
	TYPE_META_DEFAULT_COLOR    string = "#428bca"
	TYPE_META_DEFAULT_PRIORITY int    = 10
)

type IEventAnnotationTypes interface {
	GetType(string) (EventAnnoType, error)
	UpsertType(EventAnnoType) error
	RemoveType(string) error
	ListTypes() []EventAnnoType
}

type EventAnnoType struct {
	Id       string                 `json:"id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

func NewEventAnnoType(id, name string) *EventAnnoType {
	return &EventAnnoType{id, name, map[string]interface{}{
		"priority": TYPE_META_DEFAULT_PRIORITY,
		"color":    TYPE_META_DEFAULT_COLOR,
	}}
}
