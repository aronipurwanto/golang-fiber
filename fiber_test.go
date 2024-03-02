package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/mustache/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var engine = mustache.New("./template", ".mustache")

var app = fiber.New(fiber.Config{
	Views: engine,
	ErrorHandler: func(ctx *fiber.Ctx, err error) error {
		ctx.Status(fiber.StatusInternalServerError)
		return ctx.SendString("Error: " + err.Error())
	},
})

func TestRoutingHandler(t *testing.T) {
	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	})

	request := httptest.NewRequest("GET", "/", nil)
	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", string(body))
}

func TestCtx(t *testing.T) {
	app.Get("/hello", func(ctx *fiber.Ctx) error {
		name := ctx.Query("name", "Guest")
		return ctx.SendString("Hello " + name)
	})

	request := httptest.NewRequest("GET", "/hello?name=roni", nil)
	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello roni", string(body))

	request = httptest.NewRequest("GET", "/hello", nil)
	result, err = app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	body, err = io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Guest", string(body))
}

func TestHttpRequest(t *testing.T) {
	app.Get("/request", func(ctx *fiber.Ctx) error {
		first := ctx.Get("firstname")
		last := ctx.Cookies("lastname")

		return ctx.SendString("Hello " + first + " " + last)
	})

	request := httptest.NewRequest("GET", "/request", nil)
	request.Header.Set("firstname", "Roni")
	request.AddCookie(&http.Cookie{Name: "lastname", Value: "Purwanto"})

	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Roni Purwanto", string(body))
}

func TestRouteParameter(t *testing.T) {
	app.Get("/user/:userId/orders/:orderId", func(ctx *fiber.Ctx) error {
		userId := ctx.Params("userId")
		orderId := ctx.Params("orderId")
		return ctx.SendString("Get Order " + orderId + " From " + userId)
	})

	request := httptest.NewRequest("GET", "/user/1/orders /12345", nil)
	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Get Order 12345 From 1", string(body))
}

func TestFormRequest(t *testing.T) {
	app.Post("/hello", func(ctx *fiber.Ctx) error {
		name := ctx.FormValue("name")
		return ctx.SendString("Hello " + name)
	})

	body := strings.NewReader("name=Roni")
	request := httptest.NewRequest("POST", "/hello", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Roni", string(bytes))
}

//go:embed source/contoh.txt
var contohFile []byte

func TestFormUpload(t *testing.T) {
	app.Post("/upload", func(ctx *fiber.Ctx) error {
		file, err := ctx.FormFile("files")
		if err != nil {
			return err
		}

		err = ctx.SaveFile(file, "./target/"+file.Filename)
		if err != nil {
			return err
		}

		return ctx.SendString("Upload success")
	})

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	file, err := writer.CreateFormFile("files", "contoh.txt")
	assert.Nil(t, err)

	file.Write(contohFile)
	writer.Close()

	request := httptest.NewRequest("POST", "/upload", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Upload success", string(bytes))
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestRequestBody(t *testing.T) {
	app.Post("/login", func(ctx *fiber.Ctx) error {
		body := ctx.Body()

		request := new(LoginRequest)
		err := json.Unmarshal(body, request)
		if err != nil {
			return err
		}

		return ctx.SendString("Login " + request.Username + " success")
	})

	body := strings.NewReader(`{"username":"Roni","password":"rahasia"}`)
	request := httptest.NewRequest("POST", "/login", body)
	request.Header.Set("Content-Type", "application/json")

	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Login Roni success", string(bytes))
}

type RegisterRequest struct {
	Username string `json:"username" xml:"username" form:"username"'`
	Password string `json:"password" xml:"password" form:"password"`
	Name     string `json:"name" xml:"name" form:"name"`
}

func TestBodyParser(t *testing.T) {
	app.Post("/register", func(ctx *fiber.Ctx) error {
		request := new(RegisterRequest)
		err := ctx.BodyParser(request)
		if err != nil {
			return err
		}

		return ctx.SendString("Register " + request.Username + " success")
	})
}

func TestBodyParserJSON(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`{"username":"Roni","password":"rahasia","name":"Roni Purwanto"}`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/json")

	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Roni success", string(bytes))
}

func TestBodyParserForm(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`username=Roni&password=rahasia&name=Roni+Purwanto`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Roni success", string(bytes))
}

func TestBodyParserXml(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`
	<RegisterRequest>
		<username>Roni</username>
		<password>Rahasia</password>
		<name>Roni Purwanto</name>
	</RegisterRequest>
	`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/xml")

	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Roni success", string(bytes))
}

func TestResponseJSON(t *testing.T) {
	app.Get("/user", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"username": "roni",
			"name":     "Roni Purwanto",
		})
	})

	request := httptest.NewRequest("GET", "/user", nil)
	request.Header.Set("Content-Type", "application/json")

	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, `{"name":"Roni Purwanto","username":"roni"}`, string(bytes))
}

func TestDownloadFile(t *testing.T) {
	app.Get("/download", func(ctx *fiber.Ctx) error {
		return ctx.Download("./source/contoh.txt", "contoh.txt")
	})

	request := httptest.NewRequest("GET", "/download", nil)

	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, `attachment; filename="contoh.txt"`, result.Header.Get("Content-Disposition"))

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "ini contoh file yang akan di upload", string(bytes))
}

func TestRoutingGroup(t *testing.T) {
	hanlder := func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	}

	api := app.Group("/api")
	api.Get("/hello", hanlder)
	api.Get("/world", hanlder)

	web := app.Group("/web")
	web.Get("/hello", hanlder)
	web.Get("/world", hanlder)

	request := httptest.NewRequest("GET", "/api/hello", nil)
	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", string(bytes))

	request = httptest.NewRequest("GET", "/web/hello", nil)
	result, err = app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err = io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", string(bytes))
}

func TestStatic(t *testing.T) {
	app.Static("public", "./source")
	request := httptest.NewRequest("GET", "/public/contoh.txt", nil)

	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "ini contoh file yang akan di upload", string(bytes))
}

func TestErrorHandler(t *testing.T) {
	app.Get("/error", func(ctx *fiber.Ctx) error {
		return errors.New("Ups")
	})

	request := httptest.NewRequest("GET", "/error", nil)
	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 500, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Error: Ups", string(bytes))
}

func TestView(t *testing.T) {
	app.Get("/view", func(ctx *fiber.Ctx) error {
		return ctx.Render("index", fiber.Map{
			"title":   "Hello Title",
			"header":  "Hello Header",
			"content": "Hello Content",
		})
	})

	request := httptest.NewRequest("GET", "/view", nil)
	result, err := app.Test(request)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	bytes, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Contains(t, string(bytes), "Hello Title")
	assert.Contains(t, string(bytes), "Hello Header")
	assert.Contains(t, string(bytes), "Hello Content")
}

func TestClient(t *testing.T) {
	client := fiber.AcquireClient()
	defer fiber.ReleaseClient(client)

	agent := client.Get("https://example.com")
	status, response, err := agent.String()
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Contains(t, response, "Example Domain")
}
