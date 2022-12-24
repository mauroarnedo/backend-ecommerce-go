package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mauroarnedo/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var products models.Product
		defer cancel()
		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		products.Product_ID = primitive.NewObjectID()
		products.Comments = make([]models.Comment, 0)
		_, err := ProductCollection.InsertOne(ctx, products)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "not inserted"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Successfully added")
	}
}

func ProductViewerAdminBulk() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var products []models.Product
		defer cancel()

		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		productsInterface := make([]interface{}, len(products))

		for i := range products {
			products[i].Product_ID = primitive.NewObjectID()
			products[i].Comments = make([]models.Comment, 0)
			productsInterface[i] = products[i]
		}

		_, err := ProductCollection.InsertMany(ctx, productsInterface)
		if err != nil {
			log.Fatal(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": "products cannot inserted"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Successfully added")
	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {

		var productList []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := ProductCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "something went wrong, please try after some time")
			return
		}

		err = cursor.All(ctx, &productList)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		if err := cursor.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid")
			return
		}
		defer cancel()
		c.IndentedJSON(200, productList)
	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("name")

		var searchProducts []models.Product
		if query == "" {
			log.Println("Query is empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid search index"})
			c.Abort()
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := ProductCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": query, "$options": "i"}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "something went wrong, please try after some time")
			return
		}

		err = cursor.All(ctx, &searchProducts)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid")
			return
		}
		defer cancel()
		c.IndentedJSON(200, searchProducts)
	}
}

func DeleteProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		_, err = ProductCollection.DeleteOne(ctx, bson.M{"_id": productID})
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			panic(err)
		}
		defer cancel()

		update := bson.M{"$pull": bson.M{"user_favorites": bson.M{"_id": productID}}}
		_, err = UserCollection.UpdateMany(ctx, bson.M{}, update)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			panic(err)
		}
		defer cancel()

		c.IndentedJSON(200, "Product was successfully deleted")
	}
}
