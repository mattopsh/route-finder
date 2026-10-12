package osrm

type routeResponse struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Routes  []routeResult `json:"routes"`
}

type routeResult struct {
	Duration float64 `json:"duration"`
	Distance float64 `json:"distance"`
}

func (r routeResponse) describe() string {
	if r.Message == "" {
		return "code=" + r.Code
	}
	return "code=" + r.Code + " message=" + r.Message
}
