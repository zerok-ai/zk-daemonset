package zkclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	types "zerok.ai/deamonset/common"
)

var injectorendpoint = "http://zerok-injector.zerok-injector.svc.cluster.local:8444/sync-runtime"

type InjectorClient struct {
	ContainerResults []types.ContainerRuntime
}

func (h *InjectorClient) SyncDataWithInjector() {
	if len(h.ContainerResults) == 0 {
		fmt.Println("Len of container results is 0.Hence skipping sync.")
		return
	}
	containerResults := h.ContainerResults
	h.ContainerResults = []types.ContainerRuntime{}
	requestPayload := types.RuntimeSyncRequest{RuntimeDetails: containerResults}
	fmt.Println(requestPayload)
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(requestPayload)
	r, err := http.NewRequest("POST", injectorendpoint, bytes.NewBuffer(reqBodyBytes.Bytes()))
	if err != nil {
		panic(err)
	}
	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		//Adding back container results for next sync.
		h.ContainerResults = append(h.ContainerResults, containerResults...)
	}
}
