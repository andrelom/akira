package terminal

import "fmt"

func ClearEntireScreen() {
	fmt.Print("\x1b[2J\x1b[3J\x1b[H")
}

func ClearCharacters(total int) {
	for idx := 0; idx < total; idx++ {
		fmt.Print("\x1b[D")
		fmt.Print("\x1b[P")
	}
}
