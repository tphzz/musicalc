package logic

// The bundled scales are named after the description line of their .scl file, which is often a
// long citation. The catalog gives each one a short label, grouped and ordered for the tuning
// dropdown, that fits a phone screen. It is maintained by hand: utils/generate-scl.py only writes
// sclbundle.go and scl_resources.go, and catalog_test.go checks that both stay in sync.

// Tuning is one entry of the tuning dropdown
type Tuning struct {
	Label string // short "Group · Name" label shown to the user
	Name  string // key in AvailableScales
}

var catalog = []struct {
	filename string
	label    string
}{
	// Equal temperaments and their subsets
	{"tet-12.scl", "Equal · 12-TET"},
	{"24-60.scl", "Equal · 12+15-TET (Novaro)"},
	{"12-31.scl", "Equal · 31-TET meantone 12"},
	{"19-31.scl", "Equal · 31-TET meantone 19"},
	{"miracle1.scl", "Equal · 72-TET Miracle 21"},
	{"fokker_53.scl", "Equal · Fokker 53-tone"},
	{"carlos_alpha.scl", "Equal · Carlos Alpha"},
	{"bohlen-p.scl", "Equal · Bohlen-Pierce"},

	// Just intonation
	{"pyth_12.scl", "Just · Pythagorean"},
	{"ji_12.scl", "Just · 7-limit 12 (Rich)"},
	{"carlos_super.scl", "Just · Carlos Super Just"},
	{"ptolemy.scl", "Just · Ptolemy/Zarlino"},
	{"euler.scl", "Just · Euler monochord"},
	{"euler_diat.scl", "Just · Euler diatonic 8"},
	{"kepler1.scl", "Just · Kepler monochord 1"},
	{"kepler2.scl", "Just · Kepler monochord 2"},
	{"partch_41.scl", "Just · Partch diamond 41"},
	{"partch_39.scl", "Just · Ur-Partch 39"},
	{"wilson_31.scl", "Just · Wilson 31-tone"},

	// Historic well temperaments and meantones
	{"dudon_werckmeister3_eb.scl", "Historic · Werckmeister III (EB)"},
	{"kirnberger1.scl", "Historic · Kirnberger I"},
	{"kirnberger2.scl", "Historic · Kirnberger II"},
	{"kirnberger3.scl", "Historic · Kirnberger III"},
	{"neidhardt1.scl", "Historic · Neidhardt I"},
	{"neidhardt3.scl", "Historic · Neidhardt III"},
	{"rameau.scl", "Historic · Rameau meantone"},
	{"young.scl", "Historic · Young (1807)"},
	{"kellner.scl", "Historic · Kellner (Bach)"},
	{"barnes.scl", "Historic · Barnes (Bach)"},
	{"lehman1.scl", "Historic · Lehman Bach I"},
	{"lehman2.scl", "Historic · Lehman Bach II"},
	{"bemetzrieder2.scl", "Historic · Bemetzrieder 2"},
	{"corrette3.scl", "Historic · Corrette (1753)"},

	// Scales from other musical traditions
	{"arabic_rast_on_c.scl", "World · Arabic Rast (C)"},
	{"arabic_bayati_and_bayati-shuri_on_d.scl", "World · Arabic Bayati (D)"},
	{"arabic_huzam_on_e.scl", "World · Arabic Huzam (E)"},
	{"79-159.scl", "World · Turkish maqam 79"},
	{"dudon_byzantine.scl", "World · Byzantine Hijaz"},
	{"indian.scl", "World · Indian shruti"},
	{"indian_12.scl", "World · North Indian 12"},
	{"chin_lu.scl", "World · Chinese Lü"},
	{"chin_shierlu.scl", "World · Chinese Shi Er Lü"},
	{"hirajoshi.scl", "World · Japanese koto"},
	{"selisir.scl", "World · Bali selisir"},
	{"alves_pelog.scl", "World · Pelog (Alves)"},
	{"alves_slendro.scl", "World · Slendro (Alves)"},
}

// Tunings returns the bundled scales in dropdown order, the default (12-TET) first
func Tunings() []Tuning {
	byFile := make(map[string]string, len(AvailableScales))
	for name, info := range AvailableScales {
		byFile[info.Filename] = name
	}
	tunings := make([]Tuning, 0, len(catalog))
	for _, c := range catalog {
		if name, ok := byFile[c.filename]; ok {
			tunings = append(tunings, Tuning{Label: c.label, Name: name})
		}
	}
	return tunings
}
