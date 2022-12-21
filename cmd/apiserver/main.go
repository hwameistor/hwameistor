package main

import (
	routers "github.com/hwameistor/hwameistor/pkg/apiserver/router"

	"github.com/gin-gonic/gin"
	"github.com/hwameistor/hwameistor/pkg/apiserver/docs"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {

	r := gin.Default()
	r = routers.CollectRoute(r)

	docs.SwaggerInfo.Title = "Swagger Example API"
	docs.SwaggerInfo.Description = "This is a sample hwameistor server."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/apis/hwameistor.io/v1alpha1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	gin.SetMode(gin.ReleaseMode)
	panic(r.Run(":80"))
}
