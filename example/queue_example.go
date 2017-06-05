package main

import (
	"encoding/json"
	"io/ioutil"
	"fmt"
	_ "net/http/pprof"
	"log"
	"net/http"

	"github.com/souriki/ali_mns"
	"github.com/gogap/logs"
)

type appConf struct {
	Url             string `json:"url"`
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:8080", nil))
	}()


	conf := appConf{}

	if bFile, e := ioutil.ReadFile("app.conf"); e != nil {
		panic(e)
	} else {
		if e := json.Unmarshal(bFile, &conf); e != nil {
			panic(e)
		}
	}

	client := ali_mns.NewAliMNSClient(conf.Url,
		conf.AccessKeyId,
		conf.AccessKeySecret)

	msg := ali_mns.MessageSendRequest{
		MessageBody:  "hello <\"souriki/ali_mns\">",
		DelaySeconds: 0,
		Priority:     8}


	queueManager := ali_mns.NewMNSQueueManager(client)


	err := queueManager.CreateSimpleQueue("test")

	if err != nil && !ali_mns.ERR_MNS_QUEUE_ALREADY_EXIST_AND_HAVE_SAME_ATTR.IsEqual(err) {
		fmt.Println(err)
		return
	}

	queue := ali_mns.NewMNSQueue("test", client)

	for i := 1 ; i < 10000 ; i++ {
		_, err := queue.SendMessage(msg)

		go func() {
			fmt.Println(queue.QPSMonitor().QPS())
		}()

		if err != nil {
			fmt.Println(err)
		} else {
			// logs.Pretty("response:", ret)
		}

		endChan := make(chan int)
		respChan := make(chan ali_mns.MessageReceiveResponse)
		errChan := make(chan error)
		go func() {
			select {
			case resp := <-respChan:
				{
					// logs.Pretty("response:", resp)
					logs.Debug("change the visibility: ", resp.ReceiptHandle)
					if ret, e := queue.ChangeMessageVisibility(resp.ReceiptHandle, 5); e != nil {
						fmt.Println(e)
					} else {
						// logs.Pretty("visibility changed", ret)
						logs.Debug("delete it now: ", ret.ReceiptHandle)
						if e := queue.DeleteMessage(ret.ReceiptHandle); e != nil {
							fmt.Println(e)
						}
						endChan <- 1
					}
				}
			case err := <-errChan:
				{
					fmt.Println(err)
					endChan <- 1
				}
			}
		}()

		queue.ReceiveMessage(respChan, errChan, 30)
		<-endChan
	}
}
