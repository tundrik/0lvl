package inspector

import (
	"reflect"
	"testing"

	"github.com/romshark/jscan/v2"
)

func TestIspector_Audit(t *testing.T) {
	type fields struct {
		parser *jscan.Parser[string]
		schema map[string]value
	}
	type args struct {
		box MsgBox
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   MsgBox
	}{
		
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := Ispector{
				parser: tt.fields.parser,
				schema: tt.fields.schema,
			}
			if got := sp.Audit(tt.args.box); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ispector.Audit() = %v, want %v", got, tt.want)
			}
		})
	}
}
