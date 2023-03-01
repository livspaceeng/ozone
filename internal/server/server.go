package server

import (
	"github.com/livspaceeng/ozone/configs"
	"github.com/livspaceeng/ozone/internal/utils"
) 

func Init() {
	utils.Init()
	configs.Init()
	r := NewRouter()
	r.Run(configs.GetConfig().GetString("server.address"))
}
