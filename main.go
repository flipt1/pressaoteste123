package main

import (
	"log"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/joho/godotenv"
)

var (
	client     *mongo.Client
	collection *mongo.Collection
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Access environment variables
	mongodbURI := os.Getenv("MONGODB_URI")
	mongodbUsername := os.Getenv("MONGODB_USERNAME")
	mongodbPassword := os.Getenv("MONGODB_PASSWORD")

	// Create a Gin router
	router := gin.Default()

	// Initialize MongoDB Atlas connection
	initMongoDB(mongodbURI, mongodbUsername, mongodbPassword)

	// Specify the path to the "templates" directory relative to your project's root
	router.LoadHTMLGlob("./templates/*")

	// Define routes
	router.GET("/", showForm)
	router.POST("/submit", submitForm)
	router.GET("/data", displayData)

	// Define a route for the user dashboard
	router.GET("/dashboard", userDashboard)
	router.POST("/dashboard/submit", submitBloodPressureData)

	// Start the server
	router.Run(":8080") // You can change the port as needed
}

// Initialize MongoDB connection
func initMongoDB(uri, username, password string) {
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.Auth = &options.Credential{
		Username: username,
		Password: password,
	}
	client, err := mongo.Connect(nil, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set the collection to use
	collection = client.Database("petri_dish").Collection("patients")
}

// Handler for displaying the form
func showForm(c *gin.Context) {
	// Render the HTML form template
	c.HTML(http.StatusOK, "form.html", gin.H{})
}

// Handler for form submission
func submitForm(c *gin.Context) {
	// Parse form data
	fullName := c.PostForm("full_name")
	email := c.PostForm("email")
	cpf := c.PostForm("cpf")

	// Insert data into MongoDB
	_, err := collection.InsertOne(nil, bson.M{
		"full_name": fullName,
		"email":     email,
		"cpf":       cpf,
	})
	if err != nil {
		log.Println(err)
	}

	// Redirect to a success page or display a confirmation message
	c.HTML(http.StatusOK, "success.html", gin.H{
		"message": "Cadastro feito com sucesso!", // You can customize this message as needed
	})
}

// Handler for displaying data from MongoDB
func displayData(c *gin.Context) {
	// Query data from MongoDB
	cursor, err := collection.Find(nil, bson.M{})
	if err != nil {
		log.Println(err)
	}
	defer cursor.Close(nil)

	// Iterate over the documents and collect the data
	var data []bson.M
	for cursor.Next(nil) {
		var document bson.M
		if err := cursor.Decode(&document); err != nil {
			log.Println(err)
		}
		data = append(data, document)
	}

	// Render an HTML template to display the data
	c.HTML(http.StatusOK, "data.html", gin.H{"data": data})
}

// Handler for the user dashboard
func userDashboard(c *gin.Context) {
    // Query blood pressure data from MongoDB
    cursor, err := collection.Find(nil, bson.M{})
    if err != nil {
        log.Println(err)
    }
    defer cursor.Close(nil)

    // Iterate over the documents and collect the data
    var bloodPressureData []struct {
        Date             string `bson:"date"`
        SystolicPressure string `bson:"systolic_pressure"`
        DiastolicPressure string `bson:"diastolic_pressure"`
    }
    for cursor.Next(nil) {
        var document struct {
            Date             string `bson:"date"`
            SystolicPressure string `bson:"systolic_pressure"`
            DiastolicPressure string `bson:"diastolic_pressure"`
        }
        if err := cursor.Decode(&document); err != nil {
            log.Println(err)
        }
        bloodPressureData = append(bloodPressureData, document)
    }

    // Render the user dashboard template with the blood pressure data
    c.HTML(http.StatusOK, "dashboard.html", gin.H{
        "bloodPressureData": bloodPressureData,
    })
}

// Handler for blood pressure data submission
func submitBloodPressureData(c *gin.Context) {
    // Parse the form data (systolic and diastolic pressure)
    systolicPressure := c.PostForm("systolicPressure")
    diastolicPressure := c.PostForm("diastolicPressure")

    // Insert data into MongoDB with separate fields for systolic and diastolic pressure
    _, err := collection.InsertOne(nil, bson.M{
        "systolic_pressure": systolicPressure,
        "diastolic_pressure": diastolicPressure,
    })
    if err != nil {
        log.Println(err)
    }

    // Redirect back to the user dashboard
    c.Redirect(http.StatusSeeOther, "/dashboard")
}


