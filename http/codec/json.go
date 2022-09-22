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

func StdReplyEncoder(reply interface{}) ([]byte, error) {
	bs, err := json.Marshal(reply)
	if err != nil {
		return nil, err
	}

	return bs, nil
}
