package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Novel struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Content string             `bson: "content"`
	Title   string             `bson:"title"`
	Author  string             `bson:"author"`
}

func main() {
	//w89EPNlkYf0tv1Er
	router := gin.Default()
	client, err := connectDB()
	collection := client.Database("novel").Collection("articles")
	if err != nil {
		defer disconnectDB(client)
	}
	router.POST("/post", createArticle(collection))
	router.GET("/novel/:id", getNovelByID(collection))
	router.PUT("/novel/:id", editArticleById(collection))
	router.Run()
}

func connectDB() (*mongo.Client, error) {
	clientOption := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOption)
	if err != nil {
		log.Println("Cannot connect to database")
		return nil, err
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Println("Cannot ping to server")
		return nil, err
	}
	return client, nil
}

func disconnectDB(client *mongo.Client) {
	if client == nil {
		return
	}
	err := client.Disconnect(context.TODO())
	if err != nil {
		log.Fatalln(err)
	}
}

func getNovelByID(collection *mongo.Collection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var novel Novel
		id, err := primitive.ObjectIDFromHex(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := collection.FindOne(context.Background(), bson.M{"id": id}).Decode(&novel); err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, novel)
	}
}

func createArticle(collection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		var novel Novel
		if err := c.ShouldBind(&novel); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := collection.InsertOne(context.Background(), novel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": result.InsertedID})
	}
}

func editArticleById(collection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			log.Fatalln("Wrong id")
		}
		var updatedArticle Novel
		if err := c.ShouldBindJSON(&updatedArticle); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error: ":err.Error()})
			return
		}
		filter := bson.M{"_id":id}
		update := bson.M{"$set":bson.M{"title": updatedArticle.Title, "content": updatedArticle.Content, "author": updatedArticle.Author}}
		result, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error: ": err.Error()})
			return
		}
		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error: ": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"Article status: ":"updated"})
	}
}
