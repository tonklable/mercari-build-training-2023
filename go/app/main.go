package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

type ItemDetail struct {
	Name          string `json:"name"`
	Category      string `json:"category"`
	ImageFilename string `json:"image"`
}

type Items struct {
	Items []ItemDetail `json:"items"`
}

type Response struct {
	Items   []ItemDetail `json:"items,omitempty"`
	Message string       `json:"message,omitempty"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	var newItem ItemDetail

	// Get image file path
	imgFilePath := c.FormValue("image")
	hash, err := calculateImageHash(imgFilePath)
	if err != nil {
		c.Logger().Errorf("Error calculating image hash: %s", err)
		res := Response{Message: "Error calculating image hash"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	// Create new item
	newItem = ItemDetail{
		Name:          c.FormValue("name"),
		Category:      c.FormValue("category"),
		ImageFilename: hash,
	}

	// Add new item to existing items
	existingItems, err := loadItemsFromJSON()
	if err != nil {
		c.Logger().Errorf("Error loading items from JSON: %s", err)
		res := Response{Message: "Error loading items from JSON"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	existingItems.Items = append(existingItems.Items, newItem)

	err = saveItemToJSON(existingItems)
	if err != nil {
		c.Logger().Errorf("Error saving items to JSON: %s", err)
		res := Response{Message: "Error saving items to JSON"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	c.Logger().Infof("Receive item: %s", newItem)
	res := Response{Items: existingItems.Items}
	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	jsonFile, err := os.Open("items.json")
	if err != nil {
		c.Logger().Errorf("Error opening items.json: %s", err)
		res := Response{Message: "Error opening items.json"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		c.Logger().Errorf("Error reading items.json: %s", err)
		res := Response{Message: "Error reading items.json"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	var jsonContent Response
	json.Unmarshal(jsonData, &jsonContent)
	return c.JSON(http.StatusOK, jsonContent)
}

func getItemDetail(c echo.Context) error {
	// Get item ID
	itemId := c.Param("itemId")

	// Change string to int
	itemIdInt, err := strconv.Atoi(itemId)
	if err != nil {
		c.Logger().Errorf("Error converting item ID to int: %s", err)
		res := Response{Message: "Error converting item ID to int"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	// Get items from JSON file
	jsonFile, err := os.Open("items.json")
	if err != nil {
		c.Logger().Errorf("Error opening items.json: %s", err)
		res := Response{Message: "Error opening items.json"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		c.Logger().Errorf("Error reading items.json: %s", err)
		res := Response{Message: "Error reading items.json"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	var jsonContent Response
	json.Unmarshal(jsonData, &jsonContent)
	if itemIdInt-1 >= 0 && itemIdInt-1 < len(jsonContent.Items) {
		ItemDetail := jsonContent.Items[itemIdInt-1]
		return c.JSON(http.StatusOK, ItemDetail)
	} else {
		res := Response{Message: "Item not found"}
		return c.JSON(http.StatusNotFound, res)
	}

}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func loadItemsFromJSON() (Items, error) {
	// Read JSON file
	data, err := os.ReadFile("items.json")

	//Parse JSON data
	var jsonItems Items
	_ = json.Unmarshal(data, &jsonItems)

	return jsonItems, err
}

func saveItemToJSON(jsonItems Items) error {
	// Save data to JSON file
	data, err := json.Marshal(jsonItems)
	_ = os.WriteFile("items.json", data, 0644)
	return err
}

func calculateImageHash(imageFilePath string) (string, error) {
	//Read image file
	imageData, err := os.ReadFile(imageFilePath)

	//Calculate hash
	hash := sha256.Sum256(imageData)

	//Convert hash to string
	hashString := hex.EncodeToString(hash[:]) + ".jpg"

	return hashString, err
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.GET("/items", getItems)
	e.GET("/items/:itemId", getItemDetail)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
