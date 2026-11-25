package mapstructure

import (
	"reflect"
	"testing"
)

func TestIssue196StructPtrSliceDeepMap(t *testing.T) {
	type WeaponState struct {
		ID int `mapstructure:"id"`
	}
	type Player struct {
		WeaponsSlice    []WeaponState  `mapstructure:"weapons-slice,deep"`
		WeaponsPtrSlice []*WeaponState `mapstructure:"weapons-ptr-slice,deep"`
	}

	source := Player{
		WeaponsSlice: []WeaponState{{ID: 1}},
		WeaponsPtrSlice: []*WeaponState{
			{ID: 2},
			nil,
		},
	}

	var result map[string]any
	if err := Decode(source, &result); err != nil {
		t.Fatalf("got error: %s", err)
	}

	expected := map[string]any{
		"weapons-slice": []map[string]any{{"id": 1}},
		"weapons-ptr-slice": []map[string]any{
			{"id": 2},
			nil,
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("failed:\nexpected: %#v\nresult: %#v", expected, result)
	}
}
