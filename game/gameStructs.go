package game

type Game struct {
	Players []*PlayerInGame
	Deck    *Deck
}

type PlayerInGame struct {
	ID      string
	Name    string
	Hand    []Card
	Figures []Figure
}

type Deck struct {
	Cards []Card
}

type Figure struct {
	ID       int
	Status   string
	Position int
}

type Card struct {
	ID    int
	Type  string
	Moves int
}
