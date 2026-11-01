package logic

import "fmt"

// PresetErrorKind identifies a validation problem independently of display language.
type PresetErrorKind int

const (
	PresetFileInvalid PresetErrorKind = iota
	PresetKindInvalid
	PresetVersionUnsupported
	PresetNameMissing
	PresetReferenceInvalid
	PresetMicInvalid
	PresetNotFound
	PresetNameEmpty
	PresetNameTaken
)

// PresetError retains values needed for both diagnostics and localized feedback.
type PresetError struct {
	Kind      PresetErrorKind
	Name, Mic string
	Version   int
}

func (e *PresetError) Error() string {
	switch e.Kind {
	case PresetFileInvalid:
		return "the file is not a MusiCalc preset file"
	case PresetKindInvalid:
		return "the file does not contain Alignment Delay presets"
	case PresetVersionUnsupported:
		return fmt.Sprintf("the file was made by a newer MusiCalc (preset format %d)", e.Version)
	case PresetNameMissing:
		return "a preset in the file has no name"
	case PresetReferenceInvalid:
		return fmt.Sprintf("preset “%s” has an invalid reference distance", e.Name)
	case PresetMicInvalid:
		return fmt.Sprintf("preset “%s” has an invalid distance for “%s”", e.Name, e.Mic)
	case PresetNotFound:
		return "no such preset"
	case PresetNameEmpty:
		return "enter a name"
	case PresetNameTaken:
		return fmt.Sprintf("a preset called “%s” already exists", e.Name)
	default:
		return "invalid preset"
	}
}
