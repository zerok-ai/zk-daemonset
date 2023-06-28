package main

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"zk-daemonset/internal/config"
	"zk-daemonset/internal/detector"
)

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

	fmt.Printf("Hello from daemonset\n")

	// read configuration from the file and environment variables
	cfg, err := config.ProcessArgs()
	if err != nil {
		panic(err)
	}

	fmt.Printf("config redis host:%s \n", cfg.Redis.Host)

	// start business logic
	if err := detector.Start(*cfg); err != nil {
		panic(err)
	}

	app := newApp()
	config := iris.WithLogLevel("error")
	app.Listen(":"+cfg.Server.Port, config)
}
