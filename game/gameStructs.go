package game

type Game struct {
	Players []*Player
	Deck    *Deck
}

type Player struct {
	ID      int
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
