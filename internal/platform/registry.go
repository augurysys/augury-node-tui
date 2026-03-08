package platform

type Platform struct {
	ID            string
	ScriptRelPath string
	OutputRelPath string
}

var platforms = []Platform{
	{"node2", "scripts/devices/node2-build.sh", "pkg/cl-sbc-iot-imx7"},
	{"moxa-uc3100", "scripts/devices/moxa-uc3100-build.sh", "pkg/moxa-uc31xx"},
	{"moxa-low-rpm", "scripts/devices/moxa-low-rpm-build.sh", "pkg/moxa-low-rpm"},
	{"cassia-x2000", "scripts/devices/cassia-x2000-build.sh", "pkg/cassia-x2000"},
	{"mp255-ulrpm", "scripts/devices/mp255-ulrpm.sh", "pkg/mp255-ulrpm"},
}

func Registry() []Platform {
	return append([]Platform{}, platforms...)
}

func ByID(id string) (Platform, bool) {
	for _, p := range platforms {
		if p.ID == id {
			return p, true
		}
	}
	return Platform{}, false
}
