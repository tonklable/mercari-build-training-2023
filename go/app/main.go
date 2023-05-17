package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
)

type ItemDetail struct {
	Id            int    `json:"id"`
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
	db, err := sql.Open("sqlite3", "../db/items.db")

	if err != nil {
		c.Logger().Errorf("Error opening database, %v", err)
		res := Response{Message: "Error opening database"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	defer db.Close()

	// Query database
	row, err := db.Query("SELECT id, name, category, image FROM items")
	if err != nil {
		c.Logger().Errorf("Error querying database, %v", err)
		res := Response{Message: "Error querying database"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer row.Close()

	//Iterate through rows and add to response struct
	var items []ItemDetail
	for row.Next() {
		var item ItemDetail
		err := row.Scan(&item.Id, &item.Name, &item.Category, &item.ImageFilename)
		if err != nil {
			c.Logger().Errorf("Error scanning row, %v", err)
			res := Response{Message: "Error scanning row"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		items = append(items, item)
	}

	// Check for errors
	if err := row.Err(); err != nil {
		res := Response{Message: "Error iterating through rows"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	// Return response
	res := Response{Items: items}
	return c.JSON(http.StatusOK, res)
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

	// Get items from database
	db, err := sql.Open("sqlite3", "../db/items.db")

	if err != nil {
		c.Logger().Errorf("Error opening database, %v", err)
		res := Response{Message: "Error opening database"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	defer db.Close()

	// Query database
	row := db.QueryRow("SELECT id, name, category, image FROM items WHERE id = ?", itemIdInt)

	var item ItemDetail

	// Scan row into item struct
	err = row.Scan(&item.Id, &item.Name, &item.Category, &item.ImageFilename)

	if err != nil {
		if err == sql.ErrNoRows {
			res := Response{Message: "Item not found"}
			return c.JSON(http.StatusNotFound, res)
		}
		c.Logger().Errorf("Error querying database, %v", err)
		res := Response{Message: "Error querying database"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, item)
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
