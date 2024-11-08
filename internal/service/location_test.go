package service

import (
	"github.com/crazybolillo/eryth/internal/db"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/crazybolillo/eryth/pkg/model"
	"reflect"
	"testing"
)

func TestLocationParseRow(t *testing.T) {
	cases := []struct {
		name string
		row  sqlc.ListLocationsRow
		want model.Location
	}{
		{
			name: "escaped",
			row: sqlc.ListLocationsRow{
				ID:        "rando",
				Endpoint:  db.Text("kiwi"),
				UserAgent: db.Text("LinphoneAndroid/5.2.5 (Galaxy S7 edge) LinphoneSDK/5.3.47 (tags/5.3.47^5E0)"),
				Uri:       db.Text("sip:kiwi@192.168.100.24:45331^3Btransport=tcp"),
			},
			want: model.Location{
				ID:        "rando",
				Endpoint:  "kiwi",
				UserAgent: "LinphoneAndroid/5.2.5 (Galaxy S7 edge) LinphoneSDK/5.3.47 (tags/5.3.47^0)",
				Address:   "192.168.100.24:45331;transport=tcp",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := locationParseRow(tt.row)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
