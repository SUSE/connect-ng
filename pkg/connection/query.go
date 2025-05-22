package connection

import "net/http"

// Adds the given request query to the existing HTTP request object.
func AddQuery(req *http.Request, query map[string]string) *http.Request {
	values := req.URL.Query()
	for n, v := range query {
		values.Add(n, v)
	}
	req.URL.RawQuery = values.Encode()

	return req
}
