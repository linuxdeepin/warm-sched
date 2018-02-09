package main

// implement idle and full events
type innerSource struct {
	IsInUser bool
}

func (innerSource) Scope() string { return "inner" }
func (s innerSource) Check(ids []string) []string {
	var ret []string
	for _, id := range ids {
		switch id {
		case "user":
			if s.IsInUser {
				ret = append(ret, id)
			}
		case "low-memory":
		}
	}
	return ret
}

func (s *innerSource) MarkUser() {
	s.IsInUser = true
}
