package main

import (
	"bidit/bidutil"
	"bidit/opensea"
	"bidit/opensea/model"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

var (
	WEI = decimal.NewFromInt(1000000000000000000) // 10 ** 18
)

type WorkerStatus int32

const (
	WorkerStatusNotReady WorkerStatus = iota
	WorkerStatusReady
	WorkerStatusRunning
	WorkerStatusPaused
	WorkerStatusStopped
	WorkerStatusError
)

func (status WorkerStatus) String() string {
	switch status {
	case WorkerStatusNotReady:
		return "Not Ready"
	case WorkerStatusReady:
		return "Ready"
	case WorkerStatusRunning:
		return "Running"
	case WorkerStatusPaused:
		return "Paused"
	case WorkerStatusStopped:
		return "Stopped"
	case WorkerStatusError:
		return "Error"
	}
	panic("invalid worker status")
}

type WorkerRequestSignal int32

const (
	WorkerRequestPause WorkerRequestSignal = iota
	WorkerRequestResume
	WorkerRequestStop
)

type WorkerResponseSignal int32

const (
	WorkerStatusChanged WorkerResponseSignal = iota
	WorkerProgressChanged
	WorkerParamsChanged
)

// Progress
type WorkerProgress struct {
	TokenID int
	Done    int
	Pass    int
	Fail    int
}

type WorkerParams struct {
	Slug     string
	MinPrice float64
	MaxPrice float64
	Duration int
	TokenID  int // Latest token id of the worker
}

func (params *WorkerParams) Validate(checkSlug bool) error {
	if params.Slug == "" && checkSlug {
		return PrintError("Collection slug can not be empty")
	}
	if params.MinPrice <= 0 || params.MaxPrice <= 0 {
		return PrintError("Price must bigger than 0")
	}
	if params.MaxPrice < params.MinPrice {
		return PrintError("minPrice must less than or equal to maxPrice")
	}
	if params.Duration < 15 {
		return PrintError("Duration must be at least 15 minutes")
	}
	if params.TokenID < 0 {
		return PrintError("Token ID must bigger than 0.")
	}
	return nil
}

type Worker struct {
	guard sync.Mutex

	params      WorkerParams      // Creation parameter
	status      WorkerStatus      // Current status of worker
	progress    WorkerProgress    // Current statistics of worker
	collection  *model.Collection // Updated concurrently after worker created
	minPrice    decimal.Decimal
	maxPrice    decimal.Decimal
	duration    time.Duration
	request     chan WorkerRequestSignal
	response    chan WorkerResponseSignal
	subscribers bidutil.AtomicInt
	err         error // If something bad happens, this error will be updated
}

func (worker *Worker) Log(a ...any) {
	print(worker.Params().Slug, "->", Prints(a...))
}

func NewWorker(params WorkerParams) (*Worker, error) {
	err := params.Validate(true)
	if err != nil {
		return nil, err
	}

	worker := &Worker{
		params:   params,
		status:   WorkerStatusNotReady,
		progress: WorkerProgress{},
		minPrice: decimal.NewFromFloat(params.MinPrice).Mul(WEI),
		maxPrice: decimal.NewFromFloat(params.MaxPrice).Mul(WEI),
		duration: time.Minute * time.Duration(params.Duration),
		request:  make(chan WorkerRequestSignal, 16),
		response: make(chan WorkerResponseSignal, 16),
	}

	go worker.FetchCollection()

	return worker, nil
}

func (worker *Worker) SetParams(params WorkerParams, checkSlug bool) error {
	err := params.Validate(checkSlug)
	if err != nil {
		return err
	}

	// We should pause worker because changing this values while running may be dangerous
	worker.PauseAndWait()

	worker.guard.Lock()

	worker.params.MinPrice = params.MinPrice
	worker.params.MaxPrice = params.MaxPrice
	worker.params.Duration = params.Duration

	worker.minPrice = decimal.NewFromFloat(params.MinPrice).Mul(WEI)
	worker.maxPrice = decimal.NewFromFloat(params.MaxPrice).Mul(WEI)
	worker.duration = time.Minute * time.Duration(params.Duration)

	worker.guard.Unlock()

	worker.Resume()

	worker.Broadcast(WorkerParamsChanged)

	return nil
}

// Blocking call, should go
func (worker *Worker) FetchCollection() {
	worker.Log("fetching collection")
	// Download collection
	collection, err := opensea.Collection(worker.Params().Slug)
	if err != nil {
		worker.SetError(err)
		return
	}
	// NOTE: I don't know anything about PrimaryAssetContracts
	if len(collection.PrimaryAssetContracts) != 1 {
		worker.SetError(PrintError("Not ERC721 (PrimaryAssetContracts length greater than 1)"))
		return
	}
	if collection.PrimaryAssetContracts[0].SchemaName != "ERC721" {
		worker.SetError(PrintError("Not ERC721 (Schema name is not ERC721)"))
		return
	}
	worker.guard.Lock()
	worker.collection = collection
	worker.guard.Unlock()
	// Only this function can set status to ready
	worker.SetStatus(WorkerStatusReady)
}

func (worker *Worker) Start() {
	go worker.Work()
}

// Don't call this one, call Start instead
func (worker *Worker) Work() {
	Assert(worker.Status() == WorkerStatusReady, "Worker is not ready but started.")

	defer worker.SetStatus(WorkerStatusStopped)

	gConfigGuard.Lock()
	privateKey, err := bidutil.NewPrivateKey(gConfig.Private.PrivateKey)
	Assert(err == nil, err)
	address := privateKey.PublicAddress() // Address of user who owns private key
	gConfigGuard.Unlock()

	// Find collection address
	worker.guard.Lock()
	collectionName := worker.collection.Name
	collectionAddress := worker.collection.PrimaryAssetContracts[0].Address
	worker.guard.Unlock()

	worker.Log("user address:", address)
	worker.Log("collection:", collectionName, "address:", collectionAddress)

	worker.SetStatus(WorkerStatusRunning)

	for {
		select {
		case sig := <-worker.request:
			switch sig {
			case WorkerRequestStop:
				return
			case WorkerRequestPause:
				worker.SetStatus(WorkerStatusPaused)
			case WorkerRequestResume:
				worker.SetStatus(WorkerStatusRunning)
			}
		default:
			if worker.Status() == WorkerStatusRunning {
				worker.MakeNextCollectionOffer(&privateKey, address, collectionAddress)
			} else {
				Assert(worker.Status() == WorkerStatusPaused, "Worker status error")
				time.Sleep(time.Millisecond) // Reduce cpu usage when paused
			}
		}
	}
}

func (worker *Worker) MakeNextCollectionOffer(privateKey *bidutil.PrivateKey, address, collectionAddress common.Address) {
	defer worker.IncrementProgress("tokenid")
	// Download asset
	asset, err := opensea.Asset(collectionAddress, worker.Progress().TokenID)
	if err != nil {
		worker.IncrementProgress("fail")
		worker.Log("ID", worker.Progress().TokenID, "failed with error:", err)
		return
	}
	// Find highest offer
	highest, maker := FindHighestOffer(asset)
	// Check if address already owns asset or highest bid
	if asset.Owner.Address.String() == address.String() || maker.String() == address.String() {
		worker.IncrementProgress("pass")
		worker.Log("ID", asset.TokenID, "passed because address already owns this asset or highest bid")
		return
	}
	// Calculate price
	price := worker.minPrice
	if price.LessThanOrEqual(highest) {
		const incrementFactor = 100000000000000 // 0.0001 ETH
		price = highest.Add(decimal.NewFromInt(incrementFactor))
	}
	// Check if price is higher than maxPrice
	if price.GreaterThan(worker.maxPrice) {
		worker.IncrementProgress("pass")
		worker.Log("ID", asset.TokenID, "passed because calculated price is higher than maximum price")
		return
	}
	// Make offer
	err = opensea.MakeOffer(asset, worker.collection, privateKey, price, worker.duration)
	if err != nil {
		worker.IncrementProgress("fail")
		worker.Log("ID", worker.Progress().TokenID, "failed with error:", err)
		return
	}
	worker.IncrementProgress("done")
	worker.Log("Done for ID", asset.TokenID, "Highest:", highest.Div(WEI), "Offer:", price.Div(WEI))
}

func (worker *Worker) WaitUntilReady() error {
	if worker.Status() == WorkerStatusNotReady {
		var err error
		worker.Subscribe(
			func(status WorkerStatus) bool {
				switch status {
				case WorkerStatusReady:
					return false
				case WorkerStatusStopped:
					err = PrintError("Worker stopped unexpectedly")
					return false
				case WorkerStatusError:
					err = PrintError("Worker encauntered with error:", worker.Error())
					return false
				}
				return true
			},
			nil,
			nil,
		)
		return err
	}
	return nil
}

func (worker *Worker) Params() WorkerParams {
	worker.guard.Lock()
	defer worker.guard.Unlock()
	return worker.params
}

func (worker *Worker) SetStatus(status WorkerStatus) {
	worker.guard.Lock()
	worker.status = status
	worker.guard.Unlock()
	worker.Broadcast(WorkerStatusChanged)
	worker.Log("status changed to", status)
}

func (worker *Worker) Status() WorkerStatus {
	worker.guard.Lock()
	defer worker.guard.Unlock()
	return worker.status
}

func (worker *Worker) IncrementProgress(name string) {
	worker.guard.Lock()
	switch name {
	case "tokenid":
		if worker.progress.TokenID < int(worker.collection.Stats.TotalSupply) {
			worker.progress.TokenID++
		} else {
			worker.progress.TokenID = 0
		}
	case "done":
		worker.progress.Done++
	case "pass":
		worker.progress.Pass++
	case "fail":
		worker.progress.Fail++
	default:
		panic("unknown worker progress field")
	}
	worker.guard.Unlock()
	worker.Broadcast(WorkerProgressChanged)
}

func (worker *Worker) Progress() WorkerProgress {
	worker.guard.Lock()
	defer worker.guard.Unlock()
	return worker.progress
}

func (worker *Worker) SetError(err error) {
	worker.guard.Lock()
	worker.err = err
	worker.guard.Unlock()
	// May be for clearing error
	if err != nil {
		worker.SetStatus(WorkerStatusError)
	}
}

func (worker *Worker) Error() error {
	worker.guard.Lock()
	defer worker.guard.Unlock()
	return worker.err
}

func (worker *Worker) Broadcast(signal WorkerResponseSignal) {
	for i := 0; i < worker.subscribers.Get(); i++ {
		worker.response <- signal
	}
}

// Blocking call
func (worker *Worker) Subscribe(onStatusChange func(status WorkerStatus) bool, onProgressChange func(progress WorkerProgress) bool, onParamsChange func(params WorkerParams) bool) {
	Assert(worker.subscribers.Get() <= 16, "Max 16 subscribers at same time!")
	Assert(onStatusChange != nil || onProgressChange != nil || onParamsChange != nil, "One of the functions must non-nil")
	worker.subscribers.Increment()
	defer worker.subscribers.Decrement()
	for {
		select {
		case resp := <-worker.response:
			switch resp {
			case WorkerStatusChanged:
				if onStatusChange != nil {
					if !onStatusChange(worker.Status()) {
						return
					}
				}
			case WorkerProgressChanged:
				if onProgressChange != nil {
					if !onProgressChange(worker.Progress()) {
						return
					}
				}
			case WorkerParamsChanged:
				if onParamsChange != nil {
					if !onParamsChange(worker.Params()) {
						return
					}
				}
			}
		}
	}
}

func (worker *Worker) WaitUntilStatusChange(reqStatus WorkerStatus) {
	worker.Subscribe(
		func(status WorkerStatus) bool {
			if status == reqStatus {
				return false // Means now break the subscribe loop and return
			}
			return true
		},
		nil,
		nil,
	)
}

func (worker *Worker) Pause() {
	if worker.Status() == WorkerStatusRunning {
		worker.request <- WorkerRequestPause
	}
}

func (worker *Worker) PauseAndWait() {
	if worker.Status() == WorkerStatusRunning {
		worker.request <- WorkerRequestPause
		worker.WaitUntilStatusChange(WorkerStatusPaused)
	}
}

func (worker *Worker) Resume() {
	if worker.Status() == WorkerStatusPaused {
		worker.request <- WorkerRequestResume
	}
}

func (worker *Worker) Stop() {
	switch worker.Status() {
	case WorkerStatusRunning, WorkerStatusPaused:
		worker.request <- WorkerRequestStop
	}
}

func (worker *Worker) StopAndWait() {
	switch worker.Status() {
	case WorkerStatusRunning, WorkerStatusPaused:
		worker.Stop()
		worker.WaitUntilStatusChange(WorkerStatusStopped)
	}
}
