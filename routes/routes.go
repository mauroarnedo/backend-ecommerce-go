package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mauroarnedo/ecommerce/controllers"
)

func UserRoutes(router *gin.Engine) {
	router.POST("/users/signup", controllers.SignUp())
	router.POST("/users/login", controllers.Login())
	router.GET("/users/userinfo", controllers.GetUser())
	router.DELETE("/users/userdelete", controllers.DeleteUser())
	router.POST("/admin/addproduct", controllers.ProductViewerAdmin())
	router.POST("/admin/addmanyproducts", controllers.ProductViewerAdminBulk())
	router.DELETE("/admin/deleteProduct", controllers.DeleteProduct())
	router.GET("/users/productview", controllers.SearchProduct())
	router.GET("/users/search", controllers.SearchProductByQuery())
	router.GET("/users/addfavorites", controllers.AddFavorite())
	router.GET("/users/removefavorites", controllers.RemoveFavorites())
}
