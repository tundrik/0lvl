package main

import (
	"crypto/rand"
	"encoding/json"
	"os"
	"strings"
	"time"
	"unsafe"

	"0lvl/internal/repository"

	"github.com/rs/zerolog"
	fake "github.com/brianvoe/gofakeit/v6"
	stan "github.com/nats-io/stan.go"
)

func createLogger() zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	return zerolog.New(output).With().Timestamp().Logger()
}

func genOrder() repository.Order {
	items := make([]repository.Item, 0)

	entry := "WBIL"
	track := entry + strings.ToUpper(nonceGenerate(16))

	orderId := nonceGenerate(16)
	customerId := nonceGenerate(16)
	orderUid := orderId + customerId

	for i := 0; i < 2; i++ {
		item := repository.Item{
			ChrtId:      fake.Number(1, 9999999),
			TrackNumber: track,
			Price:       199,
			Rid:         nonceGenerate(16) + customerId,
			Name:        fake.ProductName(),
			Sale:        0,
			Size:        "0",
			TotalPrice:  199,
			NmId:        fake.Number(1, 9999999),
			Brand:       fake.Company(),
			Status:      0,
		}
		items = append(items, item)
	}

	order := repository.Order{
		OrderUid:    orderUid,
		TrackNumber: track,
		Entry:       entry,
		Delivery: repository.Delivery{
			Name:    fake.Name(),
			Phone:   "+" + fake.PhoneFormatted(),
			Zip:     fake.Zip(),
			City:    fake.City(),
			Address: fake.Street(),
			Region:  fake.State(),
			Email:   fake.Email(),
		},
		Payment: repository.Payment{
			Transaction:  orderUid,
			RequestId:    "",
			Currency:     fake.CurrencyShort(),
			Provider:     "wbpay",
			Amount:       3000,
			PaymentDt:    0,
			Bank:         "SberBank",
			DeliveryCost: 2403,
			GoodsTotal:   597,
			CustomFee:    0,
		},
		Items:             items,
		Locale:            fake.LanguageAbbreviation(),
		InternalSignature: "fdhdhdghgfjhfjdghdgh",
		CustomerId:        customerId,
		DeliveryService:   fake.RandomString([]string{"meest", "nova poshta"}),
		Shardkey:          "9",
		SmId:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}

	
	return order
}

func main() {
	time.Local = time.UTC

	logger := createLogger()

	sc, err := stan.Connect("test-cluster", "client-2"); if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
	defer sc.Close()

	for {
		dataOrder := genOrder()

		b, _ := json.Marshal(dataOrder)

		err = sc.Publish("order", b); if err != nil {
			logger.Fatal().Err(err).Msg("")
		}

        logger.Info().Str("uid", dataOrder.OrderUid).Msg("publish order")
		time.Sleep(3000 * time.Millisecond) 
	}
}

var alphaWb = []byte("abcdefghijklmnopqrstuvwxyz123456789")

func nonceGenerate(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	for i := 0; i < size; i++ {
		b[i] = alphaWb[b[i]%byte(len(alphaWb))]
	}
	return *(*string)(unsafe.Pointer(&b))
}
