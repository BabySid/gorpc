package codec

import "encoding/json"

func StdParamsDecoder(raw interface{}, params interface{}) error {
	bs, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, params)
	return err
}
