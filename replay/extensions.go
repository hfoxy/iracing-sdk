package replay

import "slices"

var validExtensions = []string{".zsitrpy", ".gzitrpy", ".itrpy"}

func IsValidExtension(extension string) bool {
	return slices.Contains(validExtensions, extension)
}
