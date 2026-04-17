// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package importer

import models "solo/internal/collection"

type Importer interface {
	Import(path string) (*models.Collection, error)
}
