package importer

import models "solo/internal/collection"

type Importer interface {
	Import(path string) (*models.Collection, error)
}
