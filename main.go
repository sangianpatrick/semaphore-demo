package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"
)

// ErrLockedRedeem is an error
var ErrLockedRedeem = errors.New("Voucher Redemtion is Locked by another user")

// Result is a struct
type Result struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
}

// Voucher is a struct
type Voucher struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Value      int    `json:"value"`
	RedeemedBy string `json:"redeemedBy"`
	RedeemedAt string `json:"redeemedAt"`
}

func redeemVoucher(sw *semaphore.Weighted, userID string) (*Voucher, error) {
	if ok := sw.TryAcquire(1); !ok {
		return nil, ErrLockedRedeem
	}

	// err := sw.Acquire(context.Background(), 1)

	// if err != nil {
	// 	return nil, err
	// }

	time.Sleep(time.Second * 5)

	v := &Voucher{
		ID:         uuid.New().String(),
		Name:       "Voucher Makan Rp. 50.000,- di FoodCourt Lt. 3 Menara Multimedia",
		Value:      50000,
		RedeemedBy: userID,
		RedeemedAt: time.Now().Format(time.RFC3339),
	}

	sw.Release(1)

	return v, nil
}

func main() {
	SemWeighted := semaphore.NewWeighted(2)

	http.HandleFunc("/semaphore/voucher/redeem", func(w http.ResponseWriter, r *http.Request) {

		// In this case, the voucher could be redeemed by less then equals 2 users (thread) at the same time.
		// Voucher redemtion determined as an expensive task. It needs 3 seconds until the task is finished.
		// When 3 or more users (threads) are accessing the function "redeemVoucher()" at the same time, only 2 users (threads) that are allowed to acquire the task until it is finished,
		// and the others will be blocked then retrieved an error "ErrLockedRedeem"

		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method is not allowed"))
			return
		}
		result := make(chan *Result)
		userID := uuid.New().String()

		go func() {
			voucher, err := redeemVoucher(SemWeighted, userID)
			if err != nil {
				result <- &Result{
					Success: false,
					Data:    nil,
					Message: err.Error(),
					Code:    http.StatusLocked,
				}
			}

			result <- &Result{
				Success: true,
				Data:    voucher,
				Message: "Yey, vouhcer berhasil di redeem",
				Code:    http.StatusOK,
			}
		}()

		res := <-result

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
	})

	http.ListenAndServe(":8080", nil)
}
