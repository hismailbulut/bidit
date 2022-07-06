package main

import (
	"bidit/bidutil"
	"bidit/opensea"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	FirstStartDialogSize      = fyne.NewSize(400, 400)
	PasswordRequestDialogSize = fyne.NewSize(400, 200)
	CreateOrderDialogSize     = fyne.NewSize(400, 400)
	EditOrderDialogSize       = fyne.NewSize(400, 400)
)

// Send true if gPopupActive is true
var gShutdownPopupChan chan bool

// True if there is an open popup
var gPopupActive bidutil.AtomicBool

func InitPopups() {
	gShutdownPopupChan = make(chan bool)
}

func ShutdownPopups() {
	if gPopupActive.Get() {
		gShutdownPopupChan <- true
	}
}

func ShowInfoPopup(title, message string) {
	dialog.ShowInformation(title, message, gWindow)
}

type PopupItem struct {
	Label  string
	Hint   string
	Widget fyne.Widget
}

// Common function for all popups, must go because it block until popup closes without any error
func ShowEntryPopup(title, confirmText, dismissText string, size fyne.Size, items []PopupItem, onConfirm func() error, onDismiss func()) {
	// Only one popup at a time
	Assert(!gPopupActive.Get(), "Multiple popups opened at same time")
	Assert(len(items) > 0, "At least one item is required for creating a popup")

	table := container.NewGridWithColumns(2)
	for i := 0; i < len(items); i++ {
		table.Add(container.NewHBox(
			widget.NewLabel(items[i].Label),
			NewHelpIcon(items[i].Hint),
		))
		table.Add(items[i].Widget)
	}

	errorChan := make(chan error)
	onSubmit := func(b bool) {
		if b {
			errorChan <- onConfirm()
		} else {
			if onDismiss != nil {
				onDismiss()
			}
			errorChan <- nil
		}
	}

	errMsg := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{
		Bold: true,
	})
	content := container.NewVBox(
		table,
		errMsg,
	)

	popup := dialog.NewCustomConfirm(title, confirmText, dismissText, content, onSubmit, gWindow)
	popup.Resize(size)
	go popup.Show()

	gPopupActive.Set(true)
	defer gPopupActive.Set(false)

	// Focus to first entry if one available
	for i := 0; i < len(items); i++ {
		entry, ok := items[i].Widget.(*widget.Entry)
		if ok {
			Focus(entry)
			break
		}
	}

	for {
		select {
		case err := <-errorChan:
			if err == nil {
				return
			}
			errMsg.SetText(err.Error())
			popup.Show()
		case <-gShutdownPopupChan:
			return
		}
	}
}

func ShowFirstStartDialog() {
	input1 := widget.NewPasswordEntry()
	input1.Validator = PasswordValidator

	input2 := widget.NewPasswordEntry()
	input2.Validator = func(s string) error {
		if input1.Text != input2.Text {
			return PrintError("Password fields must same")
		}
		return nil
	}

	keyInput := widget.NewPasswordEntry()
	keyInput.Validator = func(s string) error {
		_, err := bidutil.NewPrivateKey(keyInput.Text)
		return err
	}

	testnet := widget.NewCheck("", func(b bool) {})

	onSubmit := func() error {
		// Create first config
		gConfigGuard.Lock()
		gConfig = DefaultConfig()
		gConfig.keyHash = bidutil.PasswordHash([]byte(input1.Text))
		gConfig.Private.PrivateKey = keyInput.Text
		err := gConfig.Save()
		gConfigGuard.Unlock()
		if err != nil {
			return PrintError("Failed to save config file:", err)
		}
		print("Config file created successfully")
		// Init opensea
		opensea.Init(testnet.Checked)
		// Done
		return nil
	}

	items := []PopupItem{
		{
			Label:  "Password",
			Hint:   "You must specify a strong password",
			Widget: input1,
		},
		{
			Label:  "Password again",
			Hint:   "Re-enter your password so we make sure you correctly entered it",
			Widget: input2,
		},
		{
			Label:  "Private key",
			Hint:   "Your account's private key. We need this for signing offer messages. We securely store your private key only in your filesystem, and the file is encrypted with a custom encryption logic which uses sha256, keccak256 and aes256 combination. Only your password can open this file. You can find your private key in your metamask account settings.",
			Widget: keyInput,
		},
		{
			Label:  "Use testnet",
			Hint:   "If this is checked program will connect to testnet api instead of mainnet. You can use this for testing.",
			Widget: testnet,
		},
	}

	go ShowEntryPopup("First Start", "Save", "Quit", FirstStartDialogSize, items, onSubmit, gWindow.Close)
}

func ShowPasswordRequestDialog() {
	input := widget.NewPasswordEntry()
	input.Validator = PasswordValidator

	testnet := widget.NewCheck("", func(b bool) {})

	onSubmit := func() error {
		var err error
		gConfigGuard.Lock()
		gConfig, err = LoadConfig(bidutil.PasswordHash([]byte(input.Text)))
		gConfigGuard.Unlock()
		if err != nil {
			return PrintError("Incorrect password")
		}
		print("Configuration loaded successfully")
		// Init opensea
		opensea.Init(testnet.Checked)
		// Load workers
		gOrderOrganizer.Load()
		// Done
		return nil
	}

	items := []PopupItem{
		{
			Label:  "Password",
			Hint:   "The password you specified when you open this program first time",
			Widget: input,
		},
		{
			Label:  "Use testnet",
			Hint:   "If this is checked program will connect to testnet api instead of mainnet. You can use this for testing.",
			Widget: testnet,
		},
	}

	go ShowEntryPopup("Enter Password", "OK", "Quit", PasswordRequestDialogSize, items, onSubmit, gWindow.Close)
}

func ShowSettingsDialog() {
	// theme := widget.NewSelectEntry([]string{"Dark", "Light"})
	// apiKey := widget.NewEntry()
	// maxRetries := widget.NewEntry()
	// cooldown := widget.NewSlider(1, 100)
	// getRate := widget.NewSlider(0, 5)
	// postRate := widget.NewSlider(0, 5)
	// widget
}

func ShowCreateOrderDialog() {
	slug := widget.NewEntry()
	slug.Validator = func(s string) error {
		if s == "" {
			return PrintError("Collection slug can not be empty")
		}
		return nil
	}

	minPrice := widget.NewEntry()
	minPrice.Validator = FloatValidator

	maxPrice := widget.NewEntry()
	maxPrice.Validator = FloatValidator

	duration := widget.NewEntry()
	duration.Validator = IntValidator

	onSubmit := func() error {
		// Check slug
		if gOrderOrganizer.IndexOf(slug.Text) != -1 {
			return PrintError("There is already an order for this collection")
		}
		// parse inputs
		worker, err := NewWorker(WorkerParams{
			Slug:     slug.Text,
			MinPrice: ConvertStringToFloat(minPrice.Text),
			MaxPrice: ConvertStringToFloat(maxPrice.Text),
			Duration: ConvertStringToInt(duration.Text),
			// TODO: Request Token id from user
		})
		if err != nil {
			return err
		}
		gOrderOrganizer.Add(worker)
		return nil
	}

	items := []PopupItem{
		{
			Label:  "Collection slug",
			Hint:   "Collection slug is the collection's path in opensea website url. You will find it at the last of the collection's opensea url.",
			Widget: slug,
		},
		{
			Label:  "Minimum price",
			Hint:   "Minimum price in ETH format",
			Widget: minPrice,
		},
		{
			Label:  "Maximum price",
			Hint:   "Maximum price in ETH format",
			Widget: maxPrice,
		},
		{
			Label:  "Duration",
			Hint:   "Duration in minutes, minimum 15",
			Widget: duration,
		},
	}

	go ShowEntryPopup("Create Order", "Create", "Cancel", CreateOrderDialogSize, items, onSubmit, nil)
}

func ShowEditOrderDialog(worker *Worker) {
	params := worker.Params()

	minPrice := widget.NewEntry()
	minPrice.Validator = FloatValidator
	minPrice.SetText(Prints(params.MinPrice))

	maxPrice := widget.NewEntry()
	maxPrice.Validator = FloatValidator
	maxPrice.SetText(Prints(params.MaxPrice))

	duration := widget.NewEntry()
	duration.Validator = IntValidator
	duration.SetText(Prints(params.Duration))

	onConfirm := func() error {
		params := WorkerParams{
			MinPrice: ConvertStringToFloat(minPrice.Text),
			MaxPrice: ConvertStringToFloat(maxPrice.Text),
			Duration: ConvertStringToInt(duration.Text),
		}
		err := worker.SetParams(params, false)
		if err != nil {
			return err
		}
		return nil
	}

	items := []PopupItem{
		{
			Label:  "Minimum price",
			Hint:   "Minimum price in ETH format",
			Widget: minPrice,
		},
		{
			Label:  "Maximum price",
			Hint:   "Maximum price in ETH format",
			Widget: maxPrice,
		},
		{
			Label:  "Duration",
			Hint:   "Duration in minutes, minimum 15",
			Widget: duration,
		},
	}

	go ShowEntryPopup("Edit Order", "OK", "Cancel", EditOrderDialogSize, items, onConfirm, nil)
}
