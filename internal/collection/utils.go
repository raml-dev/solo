package collection

import (
	"fmt"
	"slices"
)

// Utilities
func (c *Collection) exists(id string) bool {
	return slices.ContainsFunc(c.Requests,
		func(r Request) bool {
			return r.Id == id
		})
}

func (c *Collection) get(id string) (int, *Request) {
	for i, r := range c.Requests {
		if r.Id == id {
			return i, &c.Requests[i]
		}
	}
	return -1, nil
}

func (cm *CollectionManager) buildCollectionFileName(name string) string {
	// The fs directory tree will be:
	// c.fsPath (is the main path)
	// c.Name (the name of the json file containg collection)

	return fmt.Sprintf("%s/%s.json", cm.path, name)
}

func (cm *CollectionManager) collectionExists(name string) (bool, error) {
	// check if a collection with name already exists
	c, err := cm.LoadCollection(name)

	if err != nil {
		// error in reading collection with name
		return false, err
	}
	if c != nil {
		return true, nil
	}

	return false, nil
}
