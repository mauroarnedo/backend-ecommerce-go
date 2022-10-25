package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mauroarnedo/ecommerce/database"
	"github.com/mauroarnedo/ecommerce/models"
	generate "github.com/mauroarnedo/ecommerce/tokens"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	UserCollection    *mongo.Collection = database.UserData(database.Client, "users")
	ProductCollection *mongo.Collection = database.ProductData(database.Client, "products")
	Validate                            = validator.New()
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword, givenPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true
	msg := ""
	if err != nil {
		msg = "Login or Password is incorrect"
		valid = false
	}
	return valid, msg
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := Validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
			return
		}
		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is already in use"})
			return
		}
		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		token, refreshToken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
		user.Token = &token
		user.Refresh_Token = &refreshToken
		user.User_Cart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)
		_, insertErr := UserCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusCreated, "Successfully Signed Up")
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
			return
		}

		passwordValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()

		if !passwordValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}
		token, refreshToken, _ := generate.TokenGenerator(*foundUser.Email, *foundUser.First_Name, *foundUser.Last_Name, foundUser.User_ID)
		defer cancel()
		generate.UpdateAllTokens(token, refreshToken, foundUser.User_ID)
		c.JSON(http.StatusFound, foundUser)
	}
}

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
		_, err := ProductCollection.InsertOne(ctx, products)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "not inserted"})
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
