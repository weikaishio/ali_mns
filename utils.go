package ali_mns

import (
	"bytes"
	"net/http"
	"io/ioutil"

	"github.com/gogap/errors"
)

func send(client MNSClient, decoder MNSDecoder, method Method, headers map[string]string, message interface{}, resource string, v interface{}) (statusCode int, err error) {
	var resp *http.Response
	if resp, err = client.Send(method, headers, message, resource); err != nil {
		return
	}

	if resp != nil {
		defer resp.Body.Close()
		statusCode = resp.StatusCode

		if resp.StatusCode != http.StatusCreated &&
			resp.StatusCode != http.StatusOK &&
			resp.StatusCode != http.StatusNoContent {

			// get the response body
			//   the body is set in error when decoding xml failed
			bodyBytes, e := ioutil.ReadAll(resp.Body)
			if e != nil {
				err = ERR_READ_RESPONSE_BODY_FAILED.New(errors.Params{"err": e})
				return
			}

			errResp := ErrorResponse{}
			bodyReader := bytes.NewReader(bodyBytes)

			if e2 := decoder.Decode(bodyReader, &errResp); e2 != nil {
				err = ERR_UNMARSHAL_ERROR_RESPONSE_FAILED.New(errors.Params{"err": e2, "resp":string(bodyBytes)})
				return
			}
			err = ParseError(errResp, resource)
			return
		}

		if v != nil {
			if e := decoder.Decode(resp.Body, v); e != nil {
				err = ERR_UNMARSHAL_RESPONSE_FAILED.New(errors.Params{"err": e})
				return
			}
		}
	}

	return
}
