package main

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var ButtonSize = fyne.NewSize(60, 60)

type Order struct {
	guard  sync.Mutex
	worker *Worker
	gui    fyne.CanvasObject
}

func (order *Order) SetGui(content fyne.CanvasObject) {
	order.guard.Lock()
	order.gui = content
	order.guard.Unlock()
}

func (order *Order) Worker() *Worker {
	order.guard.Lock()
	defer order.guard.Unlock()
	return order.worker
}

func NewOrder(worker *Worker) *Order {
	Assert(worker != nil, "worker can not be nil")

	order := &Order{
		worker: worker,
		gui:    nil,
	}

	// Wait until order is ready
	workerError := worker.WaitUntilReady()
	if workerError != nil {
		order.CreateFailedGui(workerError)
	} else {
		order.CreateGui()
	}

	print("Order created for worker", worker.Params().Slug)

	return order
}

func (order *Order) CreateFailedGui(err error) {
	worker := order.Worker()

	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		gOrderOrganizer.Remove(worker.Params().Slug)
	})

	var retryButton *widget.Button
	retryButton = widget.NewButtonWithIcon("", theme.MediaReplayIcon(), func() {
		go worker.FetchCollection()
		retryButton.Disable()
		// Wait again until ready
		go func() {
			err := worker.WaitUntilReady()
			if err != nil {
				order.CreateFailedGui(err)
			} else {
				order.CreateGui()
			}
		}()
	})

	order.SetGui(
		NewPanel(
			container.NewVBox(
				NewCenteredLabel("Failed", false, false, true),
				NewCenteredLabel(err.Error(), true, true, false),
				container.NewCenter(
					container.NewHBox(
						NewFixed(ButtonSize, deleteButton),
						NewFixed(ButtonSize, retryButton),
					),
				),
			),
		),
	)
}

func (order *Order) CreateGui() {
	worker := order.Worker()

	worker.guard.Lock()
	Assert(worker.collection != nil, "collection can not be nil")
	imageRes, _ := fyne.LoadResourceFromURLString(worker.collection.ImageURL)
	worker.guard.Unlock()

	labelSuccess := NewCenteredLabel("0", false, true, false)
	labelPassed := NewCenteredLabel("0", false, true, false)
	labelFailed := NewCenteredLabel("0", false, true, false)
	labelTokenID := NewCenteredLabel("0", false, true, false)

	worker.guard.Lock()
	collectionName := worker.collection.Name
	totalSupply := Prints(worker.collection.Stats.TotalSupply)
	minPrice := worker.minPrice.Div(WEI).String()
	maxPrice := worker.maxPrice.Div(WEI).String()
	duration := worker.duration.String()
	worker.guard.Unlock()

	labelMinPrice := NewCenteredLabel(minPrice, false, true, false)
	labelMaxPrice := NewCenteredLabel(maxPrice, false, true, false)
	labelDuration := NewCenteredLabel(duration, false, true, false)

	startButton := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), worker.Start)

	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Are you sure?", "This action will permanently delete this order.", func(b bool) {
			if b {
				gOrderOrganizer.Remove(worker.Params().Slug)
			}
		}, gWindow)
	})

	editButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		worker.Pause()
		go ShowEditOrderDialog(worker)
	})

	// Updater
	go func() {
		defer worker.Log("gui updater stopped")
		worker.Subscribe(
			func(status WorkerStatus) bool {
				switch status {
				case WorkerStatusPaused:
					startButton.SetIcon(theme.MediaFastForwardIcon())
					startButton.OnTapped = worker.Resume
				case WorkerStatusRunning:
					startButton.SetIcon(theme.MediaPauseIcon())
					startButton.OnTapped = worker.Pause
				case WorkerStatusStopped:
					return false
				}
				// order.guard.Lock()
				// order.gui.Refresh()
				// order.guard.Unlock()
				return true
			},
			func(progress WorkerProgress) bool {
				labelSuccess.SetText(Prints(progress.Done))
				labelPassed.SetText(Prints(progress.Pass))
				labelFailed.SetText(Prints(progress.Fail))
				labelTokenID.SetText(Prints(progress.TokenID))
				// order.guard.Lock()
				// order.gui.Refresh()
				// order.guard.Unlock()
				return true
			},
			func(_ WorkerParams) bool {
				worker.guard.Lock()
				minPrice := worker.minPrice.Div(WEI).String()
				maxPrice := worker.maxPrice.Div(WEI).String()
				duration := worker.duration.String()
				worker.guard.Unlock()
				labelMinPrice.SetText(minPrice)
				labelMaxPrice.SetText(maxPrice)
				labelDuration.SetText(duration)
				return true
			},
		)
	}()

	order.SetGui(
		NewPanel(
			container.NewHBox(
				// Collection Image
				NewFixed(fyne.NewSize(200, 200),
					widget.NewIcon(imageRes),
				),
				container.NewVBox(
					// Collection Name
					NewCenteredLabel(collectionName, false, false, true),
					container.NewHBox(
						// Stats and Progress
						container.NewGridWithRows(2,
							container.NewVBox(
								NewCenteredLabel("Total Supply", false, false, true),
								NewCenteredLabel(totalSupply, false, true, false),
							),
							container.NewVBox(
								NewCenteredLabel("Token ID", false, false, true),
								labelTokenID,
							),
							container.NewVBox(
								NewCenteredLabel("Minimum Price", false, false, true),
								labelMinPrice,
							),
							container.NewVBox(
								NewCenteredLabel("Success", false, false, true),
								labelSuccess,
							),
							container.NewVBox(
								NewCenteredLabel("Maximum Price", false, false, true),
								labelMaxPrice,
							),
							container.NewVBox(
								NewCenteredLabel("Passed", false, false, true),
								labelPassed,
							),
							container.NewVBox(
								NewCenteredLabel("Duration", false, false, true),
								labelDuration,
							),
							container.NewVBox(
								NewCenteredLabel("Failed", false, false, true),
								labelFailed,
							),
						),
					),
				),
				// Buttons
				container.NewCenter(
					container.NewVBox(
						NewFixed(ButtonSize, startButton),
						NewFixed(ButtonSize, editButton),
						NewFixed(ButtonSize, deleteButton),
					),
				),
			),
		),
	)
}
