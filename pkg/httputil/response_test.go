package httputil

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestConcurrentJSONResponsesStayIsolated(t *testing.T) {
	var wg sync.WaitGroup
	errCh := make(chan error, 100)
	for id := 0; id < 100; id++ {
		id := id
		wg.Add(1)
		go func() {
			defer wg.Done()
			recorder := httptest.NewRecorder()
			WriteJSON(recorder, 200, map[string]int{"id": id})
			var payload map[string]int
			if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
				errCh <- err
				return
			}
			if payload["id"] != id {
				errCh <- fmt.Errorf("response %d contained %d", id, payload["id"])
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}
