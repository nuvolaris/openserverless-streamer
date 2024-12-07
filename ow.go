package main

import (
	"net/http"

	"github.com/apache/openwhisk-client-go/whisk"
)

func NewOpenWhiskClient(apiHost string, apiKey string, namespace string) *whisk.Client {
	client, err := whisk.NewClient(http.DefaultClient,
		&whisk.Config{
			Host:      apiHost,
			Namespace: namespace,
			AuthToken: apiKey,
		})

	if err != nil {
		panic(err)
	}

	return client
}
