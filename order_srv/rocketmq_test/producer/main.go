package producer

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

func main() {
	p, err := rocketmq.NewProducer(producer.WithNameServer([]string{"192.168.244.130:9876"}))
	if err != nil {
		panic("生成producer失败")
	}

	if err := p.Start(); err != nil {
		panic("启动producer失败" + err.Error())
	}

	res, err := p.SendSync(context.Background(), primitive.NewMessage("xlt1", []byte("this is test")))
	if err != nil {
		fmt.Println("发送失败", err)
	} else {
		fmt.Println("发送成功", res.String())
	}

	if err := p.Shutdown(); err != nil {
		panic("关闭producer失败")
	}
}
