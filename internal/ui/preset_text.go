package ui

import (
	"errors"
	"musicalc/internal/i18n"
	"musicalc/internal/logic"
	"strings"
)

func presetErrorText(err error) string {
	var detail *logic.PresetError
	if !errors.As(err, &detail) {
		return err.Error()
	}
	switch detail.Kind {
	case logic.PresetFileInvalid:
		return i18n.T("preset.invalid_file")
	case logic.PresetKindInvalid:
		return i18n.T("preset.wrong_kind")
	case logic.PresetVersionUnsupported:
		return i18n.T("preset.unsupported_version", detail.Version)
	case logic.PresetNameMissing:
		return i18n.T("preset.missing_name")
	case logic.PresetReferenceInvalid:
		return i18n.T("preset.invalid_reference", detail.Name)
	case logic.PresetMicInvalid:
		return i18n.T("preset.invalid_mic", detail.Name, detail.Mic)
	case logic.PresetNotFound:
		return i18n.T("preset.not_found")
	case logic.PresetNameEmpty:
		return i18n.T("preset.enter_name")
	case logic.PresetNameTaken:
		return i18n.T("preset.duplicate_name", detail.Name)
	default:
		return err.Error()
	}
}

func presetSummary(p logic.Preset) string {
	unit := p.DistanceUnit
	if unit == "" {
		unit = logic.UnitMeters
	}
	var parts []string
	if source := strings.TrimSpace(p.Source); source != "" {
		parts = append(parts, source)
	}
	parts = append(parts, i18n.T("preset.summary_reference", logic.FormatTrim(logic.FromMeters(p.ReferenceMeters, unit), 2), unit),
		i18n.Plural("preset.summary_mics", len(p.Mics), len(p.Mics)))
	return strings.Join(parts, " · ")
}
