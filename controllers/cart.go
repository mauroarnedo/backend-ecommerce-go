package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mauroarnedo/ecommerce/database"
	"github.com/mauroarnedo/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	productCollection *mongo.Collection
	userCollection    *mongo.Collection
}

func NewApplication(productCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		productCollection: productCollection,
		userCollection:    userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := c.Query("user_ID")
		if userQueryID == "" {
			log.Println("user id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}
		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.AddProductToCart(ctx, app.productCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(200, "Successfully added to the cart")
	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := c.Query("user_ID")
		if userQueryID == "" {
			log.Println("user id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.RemoveCartItem(ctx, app.productCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(200, "item was removed from the cart successfully")
	}
}

func (app *Application) GetItemFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQuery_id := c.Query("id")
		if userQuery_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid id"})
			c.Abort()
			return
		}
		user_id, err := primitive.ObjectIDFromHex(userQuery_id)
		if err != nil {
			c.IndentedJSON(500, "Internal server error")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var filledCart models.User
		err = app.userCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: user_id}}).Decode(&filledCart)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(500, "not found")
			return
		}

		match_filter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: user_id}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$user_cart"}}}}
		group := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$user_cart.price"}}}}}}
		cursor, err := app.userCollection.Aggregate(ctx, mongo.Pipeline{match_filter, unwind, group})
		if err != nil {
			log.Println(err)
			return
		}
		var list []bson.M
		if err = cursor.All(ctx, &list); err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		for _, json := range list {
			c.IndentedJSON(200, json["total"])
			c.IndentedJSON(200, filledCart.User_Cart)
		}
		ctx.Done()
	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("id")

		if userQueryID == "" {
			log.Panicln("user id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err := database.BuyItemFromCart(ctx, app.userCollection, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		c.IndentedJSON(200, "Successfully placed the order")
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := c.Query("user_ID")
		if userQueryID == "" {
			log.Println("user id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.InstantBuy(ctx, app.productCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(200, "Successfully placed the instant order")
	}
}
