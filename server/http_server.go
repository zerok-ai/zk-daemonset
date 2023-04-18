package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LangDetect struct {
	PodUID     string   `json:"name"`
	Containers []string `json:"cont"`
	Image      string   `json:"image"`
}

func StartServer() {
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var result LangDetect
		err := json.Unmarshal(body, &result)
		fmt.Println(result.PodUID, result.Containers)
		if err != nil {
			fmt.Println("Error unmarshaling data from request.")
			w.WriteHeader(500)
		} else {
			//detector.FindLang(result.PodUID, result.Containers, result.Image)
			w.WriteHeader(200)
			w.Write([]byte("Done"))
		}
	})
	http.Handle("/lang", h)
	http.ListenAndServe(":8127", nil)
}
