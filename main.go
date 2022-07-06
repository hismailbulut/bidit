package main

import (
	"bytes"
	"log"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hismailbulut/neoray/src/measurer"
)

const (
	NAME    = "bidit"
	VERSION = "0.0.1"
)

var (
	WindowSize = fyne.NewSize(1024, 768)
)

var (
	gWindow fyne.Window

	gConfigGuard sync.Mutex
	gConfig      *Config

	gOrderOrganizer *OrderOrganizer

	gEtherPriceLabel *widget.Label
	gLowGasLabel     *widget.Label
	gAverageGasLabel *widget.Label
	gHighGasLabel    *widget.Label
)

func main() {
	// For benchmarking
	measurer.Init()
	defer measurer.Close()

	// TODO delete in release build
	log.SetOutput(new(bytes.Buffer)) // NOTE: Disables fyne output, but buffer gets bigger every time

	fyneApp := app.New()
	// fyneApp.Settings().SetTheme(BiditTheme{})

	gWindow = fyneApp.NewWindow(NAME)
	gWindow.Resize(WindowSize)
	gWindow.CenterOnScreen()

	InitPopups()

	gOrderOrganizer = NewOrderOrganizer()

	gEtherPriceLabel = widget.NewLabel("Ethusd: 0")
	gLowGasLabel = widget.NewLabel("Low: 0")
	gAverageGasLabel = widget.NewLabel("Average: 0")
	gHighGasLabel = widget.NewLabel("High: 0")

	settingsButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() { print("settings button pressed") })

	gWindow.SetContent(
		container.NewBorder(
			container.NewCenter(
				container.NewHBox(
					gEtherPriceLabel,
					gLowGasLabel,
					gAverageGasLabel,
					gHighGasLabel,
					settingsButton,
				),
			),
			nil,
			nil,
			nil,
			container.NewVScroll(
				container.NewVBox(
					gOrderOrganizer.gui,
					container.NewCenter(
						widget.NewButton("New Order", ShowCreateOrderDialog),
					),
				),
			),
		),
	)

	stopGasUpdater := make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Millisecond * 5100) // Etherscan gives 1 request per 5 seconds, we add a +100 ms for warranty
		defer ticker.Stop()
		for i := 0; true; i++ {
			select {
			case <-ticker.C:
				if i%4 == 0 {
					// We will update ethereum price in every 20.4 second
					ethprice, err := FetchEthPrice()
					if err != nil {
						print("FetchEthPrice:", err)
					} else {
						gEtherPriceLabel.SetText(Prints("Ethusd:", ethprice.Ethusd))
					}
				} else {
					gasprice, err := FetchGasPrice()
					if err != nil {
						print("FetchGasPrice:", err)
					} else {
						gLowGasLabel.SetText(Prints("Low:", gasprice.SafeGasPrice))
						gAverageGasLabel.SetText(Prints("Average:", gasprice.ProposeGasPrice))
						gHighGasLabel.SetText(Prints("High:", gasprice.FastGasPrice))
					}
				}
			case <-stopGasUpdater:
				return
			}
		}
	}()

	if !CheckConfigFolder() {
		ShowFirstStartDialog()
	} else {
		ShowPasswordRequestDialog()
	}

	// Blocking call
	gWindow.ShowAndRun()

	cleanupTimer := measurer.Measure()

	// Shutdown popup if one open available
	ShutdownPopups()

	// Shutdown gas updater
	stopGasUpdater <- true

	// Save and shutdown workers
	gOrderOrganizer.Save()
	gOrderOrganizer.Shutdown()

	cleanupTimer("cleanup")
}

func Focus(widget fyne.Focusable) {
	gWindow.Canvas().Focus(widget)
}
