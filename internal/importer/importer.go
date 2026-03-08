package importer

import models "yapla/internal/collection"

type Importer interface {
	Import(path string) (*models.Collection, error)
}
