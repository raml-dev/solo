package collections

import (
	"encoding/json"
	"fmt"
	"os"
)

func (c *Collection) SaveCollection() error { return nil }
func (c *Collection) LoadCollection() (*Collection, error) {
	file, err := os.ReadFile(c.buildCollectionFsName())

	if err != nil {
		return nil, err
	}
	var rC Collection

	err = json.Unmarshal(file, &rC)

	if err != nil {
		return nil, err
	}

	return &rC, nil
}
func (c *Collection) RemoveCollection() error {

	err := os.Remove(c.buildCollectionFsName())
	if err != nil {
		return err
	}
	return nil
}
func (c *Collection) UpdateCollection(collection *Collection) error {
	fName := c.buildCollectionFsName()

	file, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)

	if err != nil {
		return err
	}

	bytes, err := json.Marshal(c)

	if err != nil {
		return err
	}

	n, err := file.Write(bytes)

	if err != nil {
		return err
	}

	if n == 0 {
		return fmt.Errorf("nothing written in file %s", fName)
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

// utilities
func (c *Collection) buildCollectionFsName() string {
	// The fs directory tree will be:
	// c.fsPath (is the main path)
	// c.Name (the name of the json file containg collection)

	return fmt.Sprintf("%s/collections/%s.json", c.fsPath, c.Name)
}
