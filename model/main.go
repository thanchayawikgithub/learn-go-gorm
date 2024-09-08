package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	host     = "localhost"
	port     = 5432
	database = "go"
	username = "user"
	password = "password"
)

func middleware(c *fiber.Ctx) error {
	cookie := c.Cookies("access_token")
	jwtSecret := "Test"

	token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	claim := token.Claims.(jwt.MapClaims)
	fmt.Println(claim)
	return c.Next()
}

func main() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, username, password, database)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Enable color
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("failed to connect database")
	}

	// createBook(db, &Book{
	// 	Name:        "Mike",
	// 	Author:      "Mikelopster",
	// 	Price:       890,
	// 	Description: "sdsd",
	// })

	db.AutoMigrate(&Book{}, &User{})
	fmt.Println("Migrate successful")

	app := fiber.New()
	app.Use("/books", middleware)

	app.Get("/books", func(c *fiber.Ctx) error {
		return c.JSON(getBooks(db))
	})

	app.Get("/books/:id", func(c *fiber.Ctx) error {
		bookId, err := strconv.Atoi(c.Params("id"))

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		book := getBook(db, bookId)

		return c.JSON(book)
	})

	app.Post("/books", func(c *fiber.Ctx) error {
		book := new(Book)

		if err := c.BodyParser(book); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		err := createBook(db, book)

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		return c.JSON(fiber.Map{"message": "Create Book Successful"})
	})

	app.Put("/books/:id", func(c *fiber.Ctx) error {
		bookId, err := strconv.Atoi(c.Params("id"))

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		book := new(Book)

		if err := c.BodyParser(book); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		book.ID = uint(bookId)

		err = updateBook(db, book)

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		return c.JSON(fiber.Map{"message": "Update Book Successful"})
	})

	app.Delete("/books/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		err = deleteBook(db, id)

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		return c.JSON(fiber.Map{"message": "Delete Book Successful"})
	})

	app.Post("/register", func(c *fiber.Ctx) error {
		user := new(User)
		err := c.BodyParser(user)

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		err = createUser(db, user)

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		return c.JSON(fiber.Map{
			"message": "Register Successful",
		})
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		user := new(User)

		err := c.BodyParser(user)

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		token, err := login(db, user)

		if err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    token,
			Expires:  time.Now().Add(time.Hour * 72),
			HTTPOnly: true,
		})

		return c.JSON(fiber.Map{
			"message": "Login Successful",
		})
	})
	app.Listen(":8080")
	currentBooks := searchBook(db, "Mike")

	for _, book := range currentBooks {
		fmt.Println(book.ID, book.Name, book.Author, book.Price)
	}
}
