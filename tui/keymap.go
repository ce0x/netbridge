package tui

type KeyMap struct {
	Quit       string
	Up         string
	Down       string
	Enter      string
	Back       string
	NumberKeys map[int]string
}

func DefaultKeyMap() *KeyMap {
	return &KeyMap{
		Quit:  "q",
		Up:    "up",
		Down:  "down",
		Enter: "enter",
		Back:  "esc",
		NumberKeys: map[int]string{
			1: "1", 2: "2", 3: "3", 4: "4", 5: "5",
			6: "6", 7: "7", 8: "8", 9: "9", 0: "0",
		},
	}
}
