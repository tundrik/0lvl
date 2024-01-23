package inspector

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/stan.go"
	"github.com/romshark/jscan/v2"
)

const (
	maxLenData = 1024 * 3
)

type MsgBox struct {
	Uid  string
	Rang int64
	Msg  *stan.Msg
	Data []byte
	Err  error
}

type Ispector struct {
	parser *jscan.Parser[string]
	schema map[string]value
}

func New() Ispector {
	return Ispector{
		parser: jscan.NewParser[string](64),
		schema: createScheme(),
	}
}

func (sp Ispector) Audit(box MsgBox) MsgBox {

	if len(box.Data) > maxLenData {
		box.Err = errors.New("max len m Data")
		return box
	}

	iter_keys := 0

	err := sp.parser.Scan(unsafeB2S(box.Data), func(i *jscan.Iterator[string]) (err bool) {
		if k := i.Key(); k != "" {
			key := k[1 : len(k)-1]

			schemaRow, ok := sp.schema[key]
			if !ok {
				box.Err = fmt.Errorf("unregistered key found: %s", key)
				return true
			}

			if schemaRow.Type != i.ValueType() {
				box.Err = fmt.Errorf("%s: audit type expected", key)
				return true
			}

			if key == "order_uid" {
				val := i.Value()
				box.Uid = val[1 : len(val)-1]
			}

			if key == "date_created" {
				val := i.Value()
				t, err := time.Parse(time.RFC3339, val[1:len(val)-1])
				if err != nil {
					box.Err = err
					return true
				}
				box.Rang = t.UnixNano()
			}

			if i.Level() == 1 {
				iter_keys++
			}
		}
		return false
	})

	if err.IsErr() {
		box.Err = err
		return box
	}

	if iter_keys != Keys {
		box.Err = errors.New("number of keys mismatch")
		return box
	}

	return box
}
