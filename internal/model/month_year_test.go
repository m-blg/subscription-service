package model

import (
	"reflect"
	"testing"
	"time"
)

func TestMonthYear_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    time.Time
		wantErr bool
	}{
		{
			name:    "Valid date",
			input:   []byte(`"07-2025"`),
			want:    time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "Null value",
			input:   []byte(`null`),
			want:    time.Time{},
			wantErr: false,
		},
		{
			name:    "Empty string",
			input:   []byte(`""`),
			want:    time.Time{},
			wantErr: false,
		},
		{
			name:    "Invalid format",
			input:   []byte(`"2025-07"`),
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "Invalid month",
			input:   []byte(`"13-2025"`),
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var my MonthYear
			err := my.UnmarshalJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !my.Time().Equal(tt.want) {
				t.Errorf("UnmarshalJSON() got = %v, want %v", my.Time(), tt.want)
			}
		})
	}
}

func TestMonthYear_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		my      MonthYear
		want    []byte
		wantErr bool
	}{
		{
			name:    "Valid date",
			my:      MonthYear(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)),
			want:    []byte(`"07-2025"`),
			wantErr: false,
		},
		{
			name:    "Zero time",
			my:      MonthYear(time.Time{}),
			want:    []byte(`null`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.my.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %s, want %s", string(got), string(tt.want))
			}
		})
	}
}

func TestMonthYear_Scan(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name    string
		value   interface{}
		want    time.Time
		wantErr bool
	}{
		{
			name:    "Scan time.Time",
			value:   now,
			want:    now,
			wantErr: false,
		},
		{
			name:    "Scan nil",
			value:   nil,
			want:    time.Time{},
			wantErr: false,
		},
		{
			name:    "Scan invalid type (string)",
			value:   "07-2025",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var my MonthYear
			err := my.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !my.Time().Equal(tt.want) {
				t.Errorf("Scan() got = %v, want %v", my.Time(), tt.want)
			}
		})
	}
}

func TestMonthYear_Value(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name    string
		my      MonthYear
		want    interface{}
		wantErr bool
	}{
		{
			name:    "Value valid time",
			my:      MonthYear(now),
			want:    now,
			wantErr: false,
		},
		{
			name:    "Value zero time",
			my:      MonthYear(time.Time{}),
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.my.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value() got = %v, want %v", got, tt.want)
			}
		})
	}
}
