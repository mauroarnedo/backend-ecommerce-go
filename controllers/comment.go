package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mauroarnedo/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddComments() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "product id is empty"})
			c.Abort()
			return
		}

		userQueryID := c.Query("user_id")
		if userQueryID == "" {
			log.Println("user id is empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "user id is empty"})
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user_comment models.Comment
		userID, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if err = UserCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user_comment.User); err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if err = c.BindJSON(&user_comment.Comment); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		defer cancel()

		user_comment.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user_comment.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		filter := bson.M{"_id": productID}
		update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "product_comments", Value: user_comment}}}}

		_, err = ProductCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "The comment cannot be added to the product")
			return
		}
		defer cancel()

		c.IndentedJSON(200, "The comment was added successfully")
	}
}

func DeleteComments() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("comment id is empty")
			c.AbortWithStatusJSON(http.StatusBadRequest, "comment id is empty")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		filter := bson.M{"_id": productID}
		_, err = ProductCollection.DeleteOne(ctx, filter)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, "An error has occurred")
			return
		}
		defer cancel()
		c.IndentedJSON(200, "The comment was deleted")
	}
}
