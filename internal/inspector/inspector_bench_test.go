package inspector

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/valyala/fastjson"
	"github.com/xeipuuv/gojsonschema"
)

//jscan_custom   	               510321	         2371 ns/op	        431.50 MB/s	       0 B/op	         0 allocs/op 

//fastjson_custom                  155524	         8249 ns/op	          124.01 MB/s	   14168 B/op	      28 allocs/op

//json Unmarshal                   73749             16023 ns/op          63.85 MB/s        1600 B/op         42 allocs/op

//gojsonschema                      7101            151393 ns/op           6.76 MB/s       73081 B/op       1105 allocs/op



//Предположим что нам нужно сохранять заказ без сигнатуры
//ord.Del(Signature)
//box.Data = ord.MarshalTo(box.Data[:0:0])

//fastjson_custom                  101170	            11772 ns/op	      86.90 MB/s	   16768 B/op	      36 allocs/op

func BenchmarkJkjscan(b *testing.B) {
	b.StopTimer()
	sp := New()
	bb := []byte(data)
	b.StartTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		newBox := MsgBox{
			Data: bb,
		}
		newBox = sp.Audit(newBox)
		if newBox.Err != nil {
			panic(fmt.Errorf("unexpected error: %s", newBox.Err))
		}
	}

}

func BenchmarkSample(b *testing.B) {
	b.Run("jscan custom", func(b *testing.B) {
		benchmarkJkjscan(b)
	})
	b.Run("fastjson custom", func(b *testing.B) {
		benchmarkFastjson(b)
	})
	b.Run("gojsonschema", func(b *testing.B) {
		benchmarkScheme(b)
	})
}

func benchmarkJkjscan(b *testing.B) {
	b.StopTimer()
	sp := New()
	bb := []byte(data)
	b.StartTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		newBox := MsgBox{
			Data: bb,
		}
		newBox = sp.Audit(newBox)
		if newBox.Err != nil {
			panic(fmt.Errorf("unexpected error: %s", newBox.Err))
		}
	}
}


func benchmarkFastjson(b *testing.B) {
	b.StopTimer()
	bb := []byte(data)
	b.StartTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	auditor := NewAuditor2()

	for i := 0; i < b.N; i++ {
		newBox := MsgBox{
			Data: bb,
		}
        newBox = auditor.Audit(newBox)
		if newBox.Err != nil {
			panic(fmt.Errorf("unexpected error: %s", newBox.Err))
		}
	}
}

func benchmarkScheme(b *testing.B) {
	b.StopTimer()
	schemaLoader := gojsonschema.NewStringLoader(sheme)
	b.StartTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {

		dok := gojsonschema.NewStringLoader(data)
		_, err := gojsonschema.Validate(schemaLoader, dok)
		if err != nil {
			panic(err.Error())
		}
	}
}

const data = `{
	"order_uid": "b563feb7b2b84b6test",
	"track_number": "WBILMTESTTRACK",
	"entry": "WBIL",
	"delivery": {
	  "name": "Test Testov",
	  "phone": "+9720000000",
	  "zip": "2639809",
	  "city": "Kiryat Mozkin",
	  "address": "Ploshad Mira 15",
	  "region": "Kraiot",
	  "email": "test@gmail.com"
	},
	"payment": {
	  "transaction": "b563feb7b2b84b6test",
	  "request_id": "",
	  "currency": "USD",
	  "provider": "wbpay",
	  "amount": 1817,
	  "payment_dt": 1637907727,
	  "bank": "alpha",
	  "delivery_cost": 1500,
	  "goods_total": 317,
	  "custom_fee": 0
	},
	"items": [
	  {
		"chrt_id": 9934930,
		"track_number": "WBILMTESTTRACK",
		"price": 453,
		"rid": "ab4219087a764ae0btest",
		"name": "Mascaras",
		"sale": 30,
		"size": "0",
		"total_price": 317,
		"nm_id": 2389212,
		"brand": "Vivienne Sabo",
		"status": 202
	  }
	],
	"locale": "en",
	"internal_signature": "",
	"customer_id": "test",
	"delivery_service": "meest",
	"shardkey": "9",
	"sm_id": 99,
	"date_created": "2021-11-26T06:22:19Z",
	"oof_shard": "1"
  }`

const sheme = `{
	"$schema": "http://json-schema.org/draft-04/schema#",
	"type": "object",
	"properties": {
	  "order_uid": {
		"type": "string"
	  },
	  "track_number": {
		"type": "string"
	  },
	  "entry": {
		"type": "string"
	  },
	  "delivery": {
		"type": "object",
		"properties": {
		  "name": {
			"type": "string"
		  },
		  "phone": {
			"type": "string"
		  },
		  "zip": {
			"type": "string"
		  },
		  "city": {
			"type": "string"
		  },
		  "address": {
			"type": "string"
		  },
		  "region": {
			"type": "string"
		  },
		  "email": {
			"type": "string"
		  }
		},
		"required": [
		  "name",
		  "phone",
		  "zip",
		  "city",
		  "address",
		  "region",
		  "email"
		]
	  },
	  "payment": {
		"type": "object",
		"properties": {
		  "transaction": {
			"type": "string"
		  },
		  "request_id": {
			"type": "string"
		  },
		  "currency": {
			"type": "string"
		  },
		  "provider": {
			"type": "string"
		  },
		  "amount": {
			"type": "integer"
		  },
		  "payment_dt": {
			"type": "integer"
		  },
		  "bank": {
			"type": "string"
		  },
		  "delivery_cost": {
			"type": "integer"
		  },
		  "goods_total": {
			"type": "integer"
		  },
		  "custom_fee": {
			"type": "integer"
		  }
		},
		"required": [
		  "transaction",
		  "request_id",
		  "currency",
		  "provider",
		  "amount",
		  "payment_dt",
		  "bank",
		  "delivery_cost",
		  "goods_total",
		  "custom_fee"
		]
	  },
	  "items": {
		"type": "array",
		"items": [
		  {
			"type": "object",
			"properties": {
			  "chrt_id": {
				"type": "integer"
			  },
			  "track_number": {
				"type": "string"
			  },
			  "price": {
				"type": "integer"
			  },
			  "rid": {
				"type": "string"
			  },
			  "name": {
				"type": "string"
			  },
			  "sale": {
				"type": "integer"
			  },
			  "size": {
				"type": "string"
			  },
			  "total_price": {
				"type": "integer"
			  },
			  "nm_id": {
				"type": "integer"
			  },
			  "brand": {
				"type": "string"
			  },
			  "status": {
				"type": "integer"
			  }
			},
			"required": [
			  "chrt_id",
			  "track_number",
			  "price",
			  "rid",
			  "name",
			  "sale",
			  "size",
			  "total_price",
			  "nm_id",
			  "brand",
			  "status"
			]
		  }
		]
	  },
	  "locale": {
		"type": "string"
	  },
	  "internal_signature": {
		"type": "string"
	  },
	  "customer_id": {
		"type": "string"
	  },
	  "delivery_service": {
		"type": "string"
	  },
	  "shardkey": {
		"type": "string"
	  },
	  "sm_id": {
		"type": "integer"
	  },
	  "date_created": {
		"type": "string"
	  },
	  "oof_shard": {
		"type": "string"
	  }
	},
	"required": [
	  "order_uid",
	  "track_number",
	  "entry",
	  "delivery",
	  "payment",
	  "items",
	  "locale",
	  "internal_signature",
	  "customer_id",
	  "delivery_service",
	  "shardkey",
	  "sm_id",
	  "date_created",
	  "oof_shard"
	]
  }`

const (
	Uid       = "order_uid"
	Track     = "track_number"
	Entry     = "entry"
	Locale    = "locale"
	Signature = "internal_signature"
	Customer  = "customer_id"
	DelivServ = "delivery_service"
	Shardkey  = "shardkey"
	SmId      = "sm_id"
	Created   = "date_created"
	OofShard  = "oof_shard"
)

const (
	Deliv = "delivery"
	
	Name = "name"
	Phone = "phone"
	Zip = "zip"
	City = "city"
	Address = "address"
	Region = "region"
	Email = "email"
)

const (
	Pay  = "payment"

	Transaction  = "transaction"
	RequestId = "request_id"
	Currency = "currency"
	Provider = "provider"
	Amount = "amount"
	PaymentDt = "payment_dt"
	Bank = "bank"
	DeliveryCost = "delivery_cost"
	GoodsTotal = "goods_total"
	CustomFee = "custom_fee"
)

const (
	Items  = "items"

	ChrtId  = "chrt_id"
	Price = "price"
	Rid = "rid"
	Sale = "sale"
	Size = "size"
	TotalPrice = "total_price"
	NmId = "nm_id"
	Brand = "brand"
	Status = "status"
)

type Auditor2 struct {
	parser fastjson.Parser
}

func NewAuditor2() Auditor2 {
	return Auditor2{}
}

func (a Auditor2) Audit(box MsgBox) MsgBox {
	var err error

	if len(box.Data) > maxLenData {
		box.Err = errors.New("max len m Data")
		return box
	}
    
	doc, err := a.parser.ParseBytes(box.Data)
	if err != nil {
		box.Err = err
		return box
	}

	ord, err := doc.Object()
	if err != nil {
		box.Err = err
		return box
	}

	//var signature []byte
    keys, obj_keys, items, items_keys := 0, 0, 0, 0

	ord.Visit(func(k []byte, v *fastjson.Value) {
		if box.Err != nil {
			return
		}

		key := unsafeB2S(k)
		switch key {
		case Signature:
			keys++
			//signature, err = v.StringBytes()
			if err != nil {
				box.Err = err
			}

		case Uid:
			keys++
			uid, err := v.StringBytes()
			if err != nil {
				box.Err = err
			}
			box.Uid = unsafeB2S(uid)

		case Created:
			keys++
			created, err := v.StringBytes()
			if err != nil {
				box.Err = err
			}

			t, err := time.Parse(time.RFC3339, unsafeB2S(created))
			if err != nil {
				box.Err = err
			}
			box.Rang = t.UnixNano()

		case SmId:
			keys++
			if v.Type() != fastjson.TypeNumber {
				box.Err = errors.New("sm_id: audit int expected")
			}

		case Track, Entry, Locale, Customer, DelivServ, Shardkey, OofShard:
			keys++
			if v.Type() != fastjson.TypeString {
				box.Err = fmt.Errorf("%s: audit string expected", key)
			}

		case Deliv, Pay:
			keys++
			obj, err := v.Object()
			if err != nil {
				box.Err = err
			}

			obj.Visit(func(k []byte, v *fastjson.Value) {
				if box.Err != nil {
					return
				}

				key := unsafeB2S(k)
				switch key {
				case Name, Phone, Zip, City, Address, Region, Email:
					obj_keys++
					if v.Type() != fastjson.TypeString {
						box.Err = fmt.Errorf("%s: audit string expected", key)
					}

				case Transaction, RequestId, Currency, Provider, Bank:
					obj_keys++
					if v.Type() != fastjson.TypeString {
						box.Err = fmt.Errorf("%s: audit string expected", key)
					}

				case Amount, PaymentDt, DeliveryCost, GoodsTotal, CustomFee:
					obj_keys++
					if v.Type() != fastjson.TypeNumber {
						box.Err = fmt.Errorf("%s: audit int expected", key)
					}

				default:
					box.Err = fmt.Errorf("unregistered key found: %s", key)
				}
			})

		case Items:
			keys++
			arr, err := v.Array()
			if err != nil {
				box.Err = err
				return
			}

			for _, item := range arr {
				iobj, err := item.Object()
				if err != nil {
					box.Err = err
					break
				}
                items++
				iobj.Visit(func(k []byte, v *fastjson.Value) {
					if box.Err != nil {
						return
					}

					key := unsafeB2S(k)
					switch key {
					case ChrtId, Price, Sale, TotalPrice, NmId, Status:
						items_keys++
						if v.Type() != fastjson.TypeNumber {
							box.Err = fmt.Errorf("%s: audit int expected", key)
						}

					case Track, Rid, Name, Size, Brand:
						items_keys++
						if v.Type() != fastjson.TypeString {
							box.Err = fmt.Errorf("%s: audit string expected", key)
						}

					default:
						box.Err = fmt.Errorf("unregistered key found: %s", key)
					}
				})
			}

		default:
			box.Err = fmt.Errorf("unregistered key found: %s", key)
		}
	})

    if keys != Keys || obj_keys != ObjKeys || items_keys != items * ItemKeys {
		box.Err = errors.New("number of keys mismatch")
		return box
	}

	// Предположим что нам нужно сохранять заказ без сигнатуры
	// или проверить ее
	ord.Del(Signature)
	box.Data = ord.MarshalTo(box.Data[:0:0])
	//a.log.Debug().Str("signa", b2s(signature)).Msg("")

	return box
}

