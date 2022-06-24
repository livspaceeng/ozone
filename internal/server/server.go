package server

import "github.com/livspaceeng/ozone/configs"

func Init() {
	configs.Init()
	r := NewRouter()
	r.Run(configs.GetConfig().GetString("server.address"))
}
