package main

import (
	"github.com/kataras/iris/v12"
	zkConfig "github.com/zerok-ai/zk-utils-go/config"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	"zk-daemonset/internal/config"
	"zk-daemonset/internal/detector"
)

var LOG_TAG = "Main"

func healthz(ctx iris.Context) {
	ctx.WriteString("healthy")
}

func newApp() *iris.Application {
	app := iris.Default()

	crs := func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Method() == iris.MethodOptions {
			ctx.Header("Access-Control-Methods",
				"POST, PUT, PATCH, DELETE")

			ctx.Header("Access-Control-Allow-Headers",
				"Access-Control-Allow-Origin,Content-Type")

			ctx.Header("Access-Control-Max-Age",
				"86400")

			ctx.StatusCode(iris.StatusNoContent)
			return
		}

		ctx.Next()
	}
	app.UseRouter(crs)
	app.AllowMethods(iris.MethodOptions)

	// app.Get("/ping", pong).Describe("healthcheck")
	app.Get("/healthz", healthz).Describe("healthcheck")

	return app
}

func main() {

	// read configuration from the file and environment variables
	var cfg config.AppConfigs
	if err := zkConfig.ProcessArgs[config.AppConfigs](&cfg); err != nil {
		panic(err)
	}

	zklogger.Init(cfg.LogsConfig)

	zklogger.Debug(LOG_TAG, "Config redis host:%s and server port: %d ", cfg.Redis.Host, cfg.Server.Port)

	app := newApp()
	config := iris.WithLogLevel("error")
	go app.Listen(":"+cfg.Server.Port, config)

	zklogger.Info(LOG_TAG, "Started iris application")

	// start business logic
	if err := detector.Start(cfg); err != nil {
		panic(err)
	}
}
