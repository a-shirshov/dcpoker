package deck

import (
	"bytes"
	"encoding/gob"
)

func encrypt(c Card) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(c); err != nil {
		return err
	}
	return nil
}