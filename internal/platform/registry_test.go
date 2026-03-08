package platform

import (
	"reflect"
	"testing"
)

func TestRegistry_ExactMappings(t *testing.T) {
	want := []struct {
		id            string
		scriptRelPath string
		outputRelPath string
	}{
		{"node2", "scripts/devices/node2-build.sh", "pkg/cl-sbc-iot-imx7"},
		{"moxa-uc3100", "scripts/devices/moxa-uc3100-build.sh", "pkg/moxa-uc31xx"},
		{"moxa-low-rpm", "scripts/devices/moxa-low-rpm-build.sh", "pkg/moxa-low-rpm"},
		{"cassia-x2000", "scripts/devices/cassia-x2000-build.sh", "pkg/cassia-x2000"},
		{"mp255-ulrpm", "scripts/devices/mp255-ulrpm.sh", "pkg/mp255-ulrpm"},
	}

	got := Registry()
	if len(got) != len(want) {
		t.Fatalf("Registry() len = %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		p := got[i]
		if p.ID != w.id || p.ScriptRelPath != w.scriptRelPath || p.OutputRelPath != w.outputRelPath {
			t.Errorf("Registry()[%d] = %+v, want ID=%q ScriptRelPath=%q OutputRelPath=%q",
				i, p, w.id, w.scriptRelPath, w.outputRelPath)
		}
	}
}

func TestByID_Found(t *testing.T) {
	cases := []struct {
		id            string
		scriptRelPath string
		outputRelPath string
	}{
		{"node2", "scripts/devices/node2-build.sh", "pkg/cl-sbc-iot-imx7"},
		{"moxa-uc3100", "scripts/devices/moxa-uc3100-build.sh", "pkg/moxa-uc31xx"},
		{"moxa-low-rpm", "scripts/devices/moxa-low-rpm-build.sh", "pkg/moxa-low-rpm"},
		{"cassia-x2000", "scripts/devices/cassia-x2000-build.sh", "pkg/cassia-x2000"},
		{"mp255-ulrpm", "scripts/devices/mp255-ulrpm.sh", "pkg/mp255-ulrpm"},
	}
	for _, c := range cases {
		p, ok := ByID(c.id)
		if !ok {
			t.Errorf("ByID(%q) ok = false, want true", c.id)
			continue
		}
		want := Platform{ID: c.id, ScriptRelPath: c.scriptRelPath, OutputRelPath: c.outputRelPath}
		if !reflect.DeepEqual(p, want) {
			t.Errorf("ByID(%q) = %+v, want %+v", c.id, p, want)
		}
	}
}

func TestByID_NotFound(t *testing.T) {
	_, ok := ByID("nonexistent")
	if ok {
		t.Error("ByID(\"nonexistent\") ok = true, want false")
	}
}
