package annotations

type EventAnnotation struct {
	Id              string                 `json:"_id"`
	Type            string                 `json:"type"`
	Message         string                 `json:"message"`
	Tags            map[string]string      `json:"tags"`
	Data            map[string]interface{} `json:"data"`
	Timestamp       float64                `json:"timestamp"`        // seconds
	PostedTimestamp float64                `json:"posted_timestamp"` /* timestamp when submitted to backend */
}

/*
type EventAnnoConfirmation struct {
	Id string `json:"id"`
	EventAnnotation
}
*/
type EventAnnotationQuery struct {
	Types []string          `json:"types"`
	Tags  map[string]string `json:"tags"`
	Start float64           `json:"start"`
	End   float64           `json:"end"`
}

type IEventAnnotation interface {
	Query(EventAnnotationQuery, int64) ([]*EventAnnotation, error)
	Annotate(*EventAnnotation) (*EventAnnotation, error)
	Get(string, string) (*EventAnnotation, error)
}
