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
	Tonic int    // pitch class (0 = C … 11 = B) the scale's 1/1 is meant for
}

// catalogScale is one bundled scale. tonic is C unless the scale names another one; the .scl
// files themselves only give intervals above an unnamed 1/1.
type catalogScale struct {
	filename string
	name     string
	tonic    int
}

const (
	tonicC = 0
	tonicD = 2
	tonicE = 4
)

var catalog = []struct {
	group  string
	scales []catalogScale
}{
	// Equal temperaments and their subsets
	{"Equal", []catalogScale{
		{"tet-12.scl", "12-TET", tonicC},
		{"24-60.scl", "12+15-TET (Novaro)", tonicC},
		{"12-31.scl", "31-TET meantone 12", tonicC},
		{"19-31.scl", "31-TET meantone 19", tonicC},
		{"miracle1.scl", "72-TET Miracle 21", tonicC},
		{"fokker_53.scl", "Fokker 53-tone", tonicC},
		{"carlos_alpha.scl", "Carlos Alpha", tonicC},
		{"bohlen-p.scl", "Bohlen-Pierce", tonicC},
	}},
	// Just intonation
	{"Just", []catalogScale{
		{"pyth_12.scl", "Pythagorean", tonicC},
		{"ji_12.scl", "7-limit 12 (Rich)", tonicC},
		{"carlos_super.scl", "Carlos Super Just", tonicC},
		{"ptolemy.scl", "Ptolemy/Zarlino", tonicC},
		{"euler.scl", "Euler monochord", tonicC},
		{"euler_diat.scl", "Euler diatonic 8", tonicC},
		{"kepler1.scl", "Kepler monochord 1", tonicC},
		{"kepler2.scl", "Kepler monochord 2", tonicC},
		{"partch_41.scl", "Partch diamond 41", tonicC},
		{"partch_39.scl", "Ur-Partch 39", tonicC},
		{"wilson_31.scl", "Wilson 31-tone", tonicC},
	}},
	// Historic well temperaments and meantones
	{"Historic", []catalogScale{
		{"dudon_werckmeister3_eb.scl", "Werckmeister III (EB)", tonicC},
		{"kirnberger1.scl", "Kirnberger I", tonicC},
		{"kirnberger2.scl", "Kirnberger II", tonicC},
		{"kirnberger3.scl", "Kirnberger III", tonicC},
		{"neidhardt1.scl", "Neidhardt I", tonicC},
		{"neidhardt3.scl", "Neidhardt III", tonicC},
		{"rameau.scl", "Rameau meantone", tonicC},
		{"young.scl", "Young (1807)", tonicC},
		{"kellner.scl", "Kellner (Bach)", tonicC},
		{"barnes.scl", "Barnes (Bach)", tonicC},
		{"lehman1.scl", "Lehman Bach I", tonicC},
		{"lehman2.scl", "Lehman Bach II", tonicC},
		{"bemetzrieder2.scl", "Bemetzrieder 2", tonicC},
		{"corrette3.scl", "Corrette (1753)", tonicC},
	}},
	// Scales from other musical traditions
	{"World", []catalogScale{
		{"arabic_rast_on_c.scl", "Arabic Rast (C)", tonicC},
		{"arabic_bayati_and_bayati-shuri_on_d.scl", "Arabic Bayati (D)", tonicD},
		{"arabic_huzam_on_e.scl", "Arabic Huzam (E)", tonicE},
		{"79-159.scl", "Turkish maqam 79", tonicC},
		{"dudon_byzantine.scl", "Byzantine Hijaz", tonicC},
		{"indian.scl", "Indian shruti", tonicC},
		{"indian_12.scl", "North Indian 12", tonicC},
		{"chin_lu.scl", "Chinese Lü", tonicC},
		{"chin_shierlu.scl", "Chinese Shi Er Lü", tonicC},
		{"hirajoshi.scl", "Japanese koto", tonicC},
		{"selisir.scl", "Bali selisir", tonicC},
		{"alves_pelog.scl", "Pelog (Alves)", tonicC},
		{"alves_slendro.scl", "Slendro (Alves)", tonicC},
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
				tunings = append(tunings, Tuning{Group: g.group, Short: sc.name, Label: g.group + " · " + sc.name, Name: name, Tonic: sc.tonic})
			}
		}
	}
	return tunings
}
