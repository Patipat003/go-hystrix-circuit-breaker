package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/api", apiHandler)
	app.Get("/api2", apiHandler2)

	app.Listen(":8001")
}

func init() {

	hystrix.ConfigureCommand("api", hystrix.CommandConfig{
		Timeout:	500,
		RequestVolumeThreshold: 1,
		ErrorPercentThreshold: 100,
		SleepWindow: 15000,
	})

	hystrix.ConfigureCommand("api2", hystrix.CommandConfig{
		Timeout:	500,
		RequestVolumeThreshold: 1,
		ErrorPercentThreshold: 100,
		SleepWindow: 15000,
	})

	hystrixStream := hystrix.NewStreamHandler()
	hystrixStream.Start()
	go http.ListenAndServe(":8002", hystrixStream)
}

func apiHandler(c *fiber.Ctx) error {

	output := make(chan string, 1)

	hystrix.Go("api", func() error {
		
		res, err := http.Get("http://localhost:8000/api") // 3rd party api
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error fetching API")
		}
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error reading response")
		}

		msg := string(data)
		fmt.Println("Received message:", msg)

		output <- msg

		return nil
	}, func(err error) error {
		fmt.Println("Hystrix fallback triggered:", err)
		return nil
	})

	out := <-output

	return c.SendString(out)
}

func apiHandler2(c *fiber.Ctx) error {

	output := make(chan string, 1)

	hystrix.Go("api2", func() error {
		
		res, err := http.Get("http://localhost:8000/api") // 3rd party api
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error fetching API")
		}
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error reading response")
		}

		msg := string(data)
		fmt.Println("Received message:", msg)

		output <- msg

		return nil
	}, func(err error) error {
		fmt.Println("Hystrix fallback triggered:", err)
		return nil
	})

	out := <-output

	return c.SendString(out)
}