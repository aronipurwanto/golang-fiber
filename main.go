package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"
)

func main() {
	app := fiber.New(fiber.Config{
		IdleTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		Prefork:      true,
	})

	app.Use("/api", func(ctx *fiber.Ctx) error {
		fmt.Println("I 'am middleware before processing request")
		err := ctx.Next()
		fmt.Println("I 'am middleware after processing request")
		return err
	})

	app.Get("/api/hello", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello world")
	})

	if fiber.IsChild() {
		fmt.Println("I 'am is child process")
	} else {
		fmt.Println("I 'am a parent process")
	}

	err := app.Listen("localhost:3000")
	if err != nil {
		panic(err)
	}
}
