package subjects

type Subject struct {
	UUID          string `json:"uuid"`
	CanonicalName string `json:"canonicalName"`
	TmeIdentifier string `json:"tmeIdentifier,omitempty"`
	Type          string `json:"type,omitempty"`
}

type SubjectLink struct {
	ApiUrl string `json:"apiUrl"`
}
