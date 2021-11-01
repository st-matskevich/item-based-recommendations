package utils

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
)

type UID int64

func (val UID) MarshalJSON() ([]byte, error) {
	str := strconv.FormatInt(int64(val), 10)
	enc := base64.StdEncoding.EncodeToString([]byte(str))
	json, err := json.Marshal(enc)
	return json, err
}

func (val *UID) UnmarshalJSON(data []byte) error {
	var enc string
	if err := json.Unmarshal(data, &enc); err != nil {
		return err
	}

	u, err := UIDFromString(enc)
	*val = u
	return err
}

func UIDFromString(enc string) (UID, error) {
	str, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return UID(0), err
	}

	i, err := strconv.ParseInt(string(str), 10, 64)
	if err != nil {
		return UID(0), err
	}

	res := UID(i)
	return res, nil
}
