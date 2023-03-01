package utils

import (
	"github.com/livspaceeng/ozone/configs"
	client "github.com/ory/keto-client-go"
)

var (
	KetoClient		*client.APIClient
)

func Init() {
	KetoClient = createKetoReadClient()
}

func createKetoReadClient() *client.APIClient {
	configs.Init()
	readUri := configs.GetConfig().GetString("keto.read.url")
	configuration := client.NewConfiguration()
	configuration.Servers = []client.ServerConfiguration{
		{
			URL: readUri,
		},
	}
	KetoClient = client.NewAPIClient(configuration)
	return KetoClient
}

func GetKetoReadClient() *client.APIClient {
	return KetoClient
}