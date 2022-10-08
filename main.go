package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sqlx.DB

func main() {

	var err error
	db, err = sqlx.Open("sqlite3", "./db/gotest.db")
	if err != nil {
		panic(err)
	}

	app := fiber.New()

	app.Post("/signup", Signup)
	app.Post("/login", Login)
	app.Get("/hello", Hello)

	app.Listen(":8000")
}

func Signup(c *fiber.Ctx) error {

	request := SignupRequest{}
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}

	if request.Username == "" || request.Password == "" {
		return fiber.ErrUnprocessableEntity
	}

	password, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	query := "insert user (username/password) values (?,?)"
	result, err := db.Exec(query, request.Username, string(password))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	user := User{
		Id:       int(id),
		Username: request.Username,
		Password: request.Password,
	}

	return c.JSON(user)
}

func Login(c *fiber.Ctx) error {
	return nil
}

func Hello(c *fiber.Ctx) error {
	return nil
}

type User struct {
	Id       int    `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
}

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Fiber() {
	app := fiber.New(fiber.Config{
		Prefork: true,
	})

	//Middleware

	app.Use("/hello", func(c *fiber.Ctx) error {
		c.Locals("name", "bond")
		fmt.Println("before")
		c.Next()
		fmt.Println("after")
		return nil
	})
	//RequestID
	app.Use(requestid.New())
	//Cors
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "*",
		AllowHeaders: "*",
	}))
	app.Use(logger.New(logger.Config{
		TimeZone: "Asia/Bangkok",
	}))

	//Middleware

	//GET
	app.Get("/hello", func(c *fiber.Ctx) error {
		name := c.Locals("name")
		return c.SendString(fmt.Sprintf("GET: Hello %v", name))
	})

	//POST
	app.Post("/hello", func(c *fiber.Ctx) error {
		return c.SendString("Post")
	})

	//Parameters Optional
	app.Get("/hello/:name/:surname", func(c *fiber.Ctx) error {
		name := c.Params("name")
		surname := c.Params("surname")
		return c.SendString("name:" + name + " surname:" + surname)
	})

	//ParamsIns
	app.Get("hello/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return fiber.ErrBadRequest
		}
		return c.SendString(fmt.Sprintf("ID = %v", id))
	})

	//Query
	app.Get("/query", func(c *fiber.Ctx) error {
		name := c.Query("name")
		surname := c.Query("surname")
		return c.SendString("name: " + name + " surname: " + surname)
	})

	//Query
	app.Get("/query2", func(c *fiber.Ctx) error {
		person := Person{}
		c.QueryParser(&person)
		return c.JSON(person)
	})

	//Wildcards
	app.Get("/wildcard/*", func(c *fiber.Ctx) error {
		wildcard := c.Params("*")
		return c.SendString(wildcard)
	})

	//Static file
	app.Static("/", "./wwwroot", fiber.Static{
		Index:         "index.html",
		CacheDuration: time.Second * 10,
	})

	//NewError
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "content not found")
	})

	//Group
	v1 := app.Group("/v1", func(c *fiber.Ctx) error {
		c.Set("Version", "V1")
		return c.Next()
	})
	v1.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello V1")
	})

	v2 := app.Group("/v2", func(c *fiber.Ctx) error {
		c.Set("Version", "V1")
		return c.Next()
	})
	v2.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello V2")
	})

	//Mount กำหนดให้ต้องวิ่งผ่าน login ก่อน
	userApp := fiber.New()
	userApp.Get("/login", func(c *fiber.Ctx) error {
		return c.SendString("Login")
	})
	// กำหนด path ที่จะวิ่งผ่าน ถ้าจะไป user ต้องผ่าน userApp ก่อน ซึ่งจะไป Path:/login
	app.Mount("/user", userApp)

	//Server MaxConn กำหนกคนเข้าใช้ ว่าจะให้เข้าใช้ได้กี่ user
	// app.Server().MaxConnsPerIP = 1

	app.Get("/server", func(c *fiber.Ctx) error {
		time.Sleep(time.Second * 30)
		return c.SendString("server")
	})

	//Environment
	app.Get("/env", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"BaseURL":     c.BaseURL(),
			"Hostname":    c.Hostname(),
			"IP":          c.IP(),
			"IPs":         c.IPs(),
			"OriginalURL": c.OriginalURL(),
			"Path":        c.Path(),
			"Protocol":    c.Protocol(),
			"Subdomains":  c.Subdomains(),
		})
	})

	//Body
	app.Post("/body", func(c *fiber.Ctx) error {
		fmt.Printf("IsJSON : %v\n", c.Is("json"))
		fmt.Println(string(c.Body()))

		person := Person{}
		err := c.BodyParser(&person)
		if err != nil {
			return err
		}
		fmt.Println(person)
		return nil

	})

	//Body
	app.Post("/body2", func(c *fiber.Ctx) error {
		fmt.Printf("IsJSON : %v\n", c.Is("json"))
		// fmt.Println(string(c.Body()))

		data := map[string]interface{}{}
		err := c.BodyParser(&data)
		if err != nil {
			return err
		}
		fmt.Println(data)
		return nil

	})

	app.Listen(":8000")

}

type Person struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
