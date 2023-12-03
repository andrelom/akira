package terminal

import "fmt"

const (
	banner  = " █████╗ ██╗  ██╗██╗██████╗  █████╗ \n██╔══██╗██║ ██╔╝██║██╔══██╗██╔══██╗\n███████║█████╔╝ ██║██████╔╝███████║\n██╔══██║██╔═██╗ ██║██╔══██╗██╔══██║\n██║  ██║██║  ██╗██║██║  ██║██║  ██║\n╚═╝  ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚═╝  ╚═╝"
	version = "v1.0.0"
)

func PrintHeader() {
	fmt.Println(banner)
	fmt.Println(version)
}
