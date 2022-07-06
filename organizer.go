package main

import (
	"bidit/opensea"
	"sync"

	"fyne.io/fyne/v2"
)

type OrderOrganizer struct {
	guard  sync.Mutex
	orders []*Order
	gui    *fyne.Container
}

func NewOrderOrganizer() *OrderOrganizer {
	return &OrderOrganizer{
		orders: make([]*Order, 0),
		gui:    NewList(),
	}
}

func (organizer *OrderOrganizer) Add(worker *Worker) {
	Assert(worker != nil, "worker can not be nil")
	go func() {
		order := NewOrder(worker)

		organizer.guard.Lock()
		defer organizer.guard.Unlock()

		order.guard.Lock()
		defer order.guard.Unlock()

		organizer.orders = append(organizer.orders, order)
		organizer.gui.Add(order.gui)
		organizer.gui.Refresh()
	}()
}

func (organizer *OrderOrganizer) IndexOf(slug string) int {
	organizer.guard.Lock()
	defer organizer.guard.Unlock()

	index := -1
	for i, order := range organizer.orders {
		order.guard.Lock()
		if order.worker.Params().Slug == slug {
			index = i
			order.guard.Unlock()
			break
		}
		order.guard.Unlock()
	}
	return index
}

func (organizer *OrderOrganizer) Remove(slug string) {
	// Find order index
	index := organizer.IndexOf(slug)
	// Remove
	organizer.guard.Lock()
	defer organizer.guard.Unlock()

	if index != -1 {
		order := organizer.orders[index]
		order.guard.Lock()
		order.worker.Stop()
		organizer.gui.Remove(order.gui)
		organizer.orders = append(organizer.orders[:index], organizer.orders[index+1:]...)
		order.guard.Unlock()
	}
}

func (organizer *OrderOrganizer) Load() {
	gConfigGuard.Lock()
	defer gConfigGuard.Unlock()

	if opensea.IsTestnet() {
		for _, p := range gConfig.Public.TestnetWorkers {
			worker, err := NewWorker(p)
			Assert(err == nil, "Saved order params must valid")
			organizer.Add(worker)
		}
	} else {
		for _, p := range gConfig.Public.MainnetWorkers {
			worker, err := NewWorker(p)
			Assert(err == nil, "Saved order params must valid")
			organizer.Add(worker)
		}
	}
}

func (organizer *OrderOrganizer) Save() {
	organizer.guard.Lock()
	defer organizer.guard.Unlock()

	gConfigGuard.Lock()
	defer gConfigGuard.Unlock()

	if gConfig == nil {
		print("Couldn't save because config not initialized")
		return
	}

	if opensea.IsTestnet() {
		gConfig.Public.TestnetWorkers = make([]WorkerParams, 0)
		for _, order := range organizer.orders {
			order.guard.Lock()
			gConfig.Public.TestnetWorkers = append(gConfig.Public.TestnetWorkers, order.worker.Params())
			order.guard.Unlock()
		}
	} else {
		gConfig.Public.MainnetWorkers = make([]WorkerParams, 0)
		for _, order := range organizer.orders {
			order.guard.Lock()
			gConfig.Public.MainnetWorkers = append(gConfig.Public.MainnetWorkers, order.worker.Params())
			order.guard.Unlock()
		}
	}
	err := gConfig.Save()

	if err != nil {
		print("Error while saving orders:", err)
	}
}

func (organizer *OrderOrganizer) Shutdown() {
	organizer.guard.Lock()
	defer organizer.guard.Unlock()

	for _, order := range organizer.orders {
		order.guard.Lock()
		order.worker.StopAndWait()
		order.guard.Unlock()
	}
}
