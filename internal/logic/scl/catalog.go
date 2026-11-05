package logic

// The bundled scales are named after the description line of their .scl file, which is often a
// long citation. The catalog gives each one a short name in a group, ordered for the grouped tuning
// dropdown, so it fits a phone screen. It is maintained by hand: utils/generate-scl.py only writes
// sclbundle.go and scl_resources.go, and catalog_test.go checks that both stay in sync.

// Tuning is one entry of the tuning dropdown
type Tuning struct {
	Group string // submenu the tuning is listed in
	Short string // short name within its group
	Label string // "Group · Short", shown for the selected tuning
	Name  string // key in AvailableScales
}

type catalogScale struct {
	filename string
	name     string
}

var catalog = []struct {
	group  string
	scales []catalogScale
}{
	// Equal temperaments and their subsets
	{"Equal", []catalogScale{
		{"tet-12.scl", "12-TET"},
		{"24-60.scl", "12+15-TET (Novaro)"},
		{"12-31.scl", "31-TET meantone 12"},
		{"19-31.scl", "31-TET meantone 19"},
		{"miracle1.scl", "72-TET Miracle 21"},
		{"fokker_53.scl", "Fokker 53-tone"},
		{"carlos_alpha.scl", "Carlos Alpha"},
		{"bohlen-p.scl", "Bohlen-Pierce"},
	}},
	// Just intonation
	{"Just", []catalogScale{
		{"pyth_12.scl", "Pythagorean"},
		{"ji_12.scl", "7-limit 12 (Rich)"},
		{"carlos_super.scl", "Carlos Super Just"},
		{"ptolemy.scl", "Ptolemy/Zarlino"},
		{"euler.scl", "Euler monochord"},
		{"euler_diat.scl", "Euler diatonic 8"},
		{"kepler1.scl", "Kepler monochord 1"},
		{"kepler2.scl", "Kepler monochord 2"},
		{"partch_41.scl", "Partch diamond 41"},
		{"partch_39.scl", "Ur-Partch 39"},
		{"wilson_31.scl", "Wilson 31-tone"},
	}},
	// Historic well temperaments and meantones
	{"Historic", []catalogScale{
		{"dudon_werckmeister3_eb.scl", "Werckmeister III (EB)"},
		{"kirnberger1.scl", "Kirnberger I"},
		{"kirnberger2.scl", "Kirnberger II"},
		{"kirnberger3.scl", "Kirnberger III"},
		{"neidhardt1.scl", "Neidhardt I"},
		{"neidhardt3.scl", "Neidhardt III"},
		{"rameau.scl", "Rameau meantone"},
		{"young.scl", "Young (1807)"},
		{"kellner.scl", "Kellner (Bach)"},
		{"barnes.scl", "Barnes (Bach)"},
		{"lehman1.scl", "Lehman Bach I"},
		{"lehman2.scl", "Lehman Bach II"},
		{"bemetzrieder2.scl", "Bemetzrieder 2"},
		{"corrette3.scl", "Corrette (1753)"},
	}},
	// Scales from other musical traditions
	{"World", []catalogScale{
		{"arabic_rast_on_c.scl", "Arabic Rast (C)"},
		{"arabic_bayati_and_bayati-shuri_on_d.scl", "Arabic Bayati (D)"},
		{"arabic_huzam_on_e.scl", "Arabic Huzam (E)"},
		{"79-159.scl", "Turkish maqam 79"},
		{"dudon_byzantine.scl", "Byzantine Hijaz"},
		{"indian.scl", "Indian shruti"},
		{"indian_12.scl", "North Indian 12"},
		{"chin_lu.scl", "Chinese Lü"},
		{"chin_shierlu.scl", "Chinese Shi Er Lü"},
		{"hirajoshi.scl", "Japanese koto"},
		{"selisir.scl", "Bali selisir"},
		{"alves_pelog.scl", "Pelog (Alves)"},
		{"alves_slendro.scl", "Slendro (Alves)"},
	}},
}

// Tunings returns the bundled scales in dropdown order, the default (12-TET) first
func Tunings() []Tuning {
	byFile := make(map[string]string, len(AvailableScales))
	for name, info := range AvailableScales {
		byFile[info.Filename] = name
	}
	var tunings []Tuning
	for _, g := range catalog {
		for _, sc := range g.scales {
			if name, ok := byFile[sc.filename]; ok {
				tunings = append(tunings, Tuning{Group: g.group, Short: sc.name, Label: g.group + " · " + sc.name, Name: name})
			}
		}
	}
	return tunings
}
