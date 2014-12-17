package annotations

type Subscription struct {
	Types []string
	Tags  map[string]string
}

func (s *Subscription) IsSubscribedMessage(anno EventAnnoConfirmation) bool {
	if len(s.Tags) < 1 {
		return true
	}
	for k, v := range anno.Tags {
		if _, ok := s.Tags[k]; ok {
			if s.Tags[k] == v {
				return true
			}
		}
	}
	return false
}
