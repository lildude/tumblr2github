package main

func TestHandler(t *testing.T) {
	testCases := []struct {
		name           string
		webhookPayload string
		logMessage     string
	}{
		{
			name:           "empty GET",
			webhookPayload: "",
			logMessage:     "INFO: empty body",
		}
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			//t.Parallel()
			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.webhookPayload))
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handler)
			handler.ServeHTTP(rr, req)
			if !strings.Contains(fauxLog.String(), tc.logMessage) {
				t.Errorf("'%v' failed.\nGot:\n%v\nExpected:\n%v", tc.name, fauxLog.String(), tc.logMessage)
			}
		})
	}
}