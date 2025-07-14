package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/Daweenci/kartenspiel/game/handlers"
)

func main() {
	app := fiber.New()

	app.Post("/join", handlers.JoinGameHandler)
	app.Post("/start", handlers.StartGameHandler)
	app.Post("/draw", handlers.DrawCardsHandler)
	app.Post("/play", handlers.PlayCardHandler)
	app.Post("/move", handlers.MovePieceHandler)
	app.Post("/end", handlers.EndGameHandler)

	log.Fatal(app.Listen(":4000"))
}
