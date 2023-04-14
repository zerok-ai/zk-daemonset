package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	detector "zerok.ai/langdetector/detector"
)

type LangDetect struct {
	PodName   string `json:"name"`
	Container string `json:"cont"`
	Image     string `json:"image"`
}

func StartServer() {
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var result LangDetect
		err := json.Unmarshal(body, &result)
		fmt.Println(result.PodName, result.Container)
		if err != nil {
			fmt.Println("Error unmarshaling data from request.")
			w.WriteHeader(500)
		} else {
			detector.FindLang(result.PodName, []string{result.Container}, result.Image)
			w.WriteHeader(200)
			w.Write([]byte("Done"))
		}
	})
	http.Handle("/lang", h)
	http.ListenAndServe(":8127", nil)
}
