package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mauroarnedo/ecommerce/controllers"
	"github.com/mauroarnedo/ecommerce/database"
	"github.com/mauroarnedo/ecommerce/middleware"
	"github.com/mauroarnedo/ecommerce/routes"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	app := controllers.NewApplication(database.ProductData(database.Client, "products"), database.UserData(database.Client, "users"))

	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)
	router.Use(middleware.Authentication())
	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoveItem())
	router.GET("/listcart", app.GetItemFromCart())
	router.POST("/addaddress", controllers.AddAddress())
	router.PUT("/edithomeaddress", controllers.EditHomeAddress())
	router.PUT("/editworkaddress", controllers.EditWorkAddress())
	router.GET("/deleteaddresses", controllers.DeleteAddress())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())
	router.POST("/addcomments", controllers.AddComments())
	router.DELETE("/deletecomments", controllers.DeleteComments())
	log.Fatal(router.Run(":" + port))
}
