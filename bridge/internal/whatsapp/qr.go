package whatsapp

import (
	"fmt"

	"github.com/mdp/qrterminal/v3"
	"os"
)

func DisplayQR(code string) {
	fmt.Println("\n========================================")
	fmt.Println("Scan this QR code with WhatsApp to login")
	fmt.Println("========================================\n")

	config := qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	}

	qrterminal.GenerateWithConfig(code, config)

	fmt.Println("\n========================================")
	fmt.Println("Waiting for scan...")
	fmt.Println("========================================")
}
