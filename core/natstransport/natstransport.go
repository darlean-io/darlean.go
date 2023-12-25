package natstransport

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/darlean-io/darlean.go/core/wire"

	"github.com/nats-io/nats.go"
)

type NatsTransport struct {
	connection   *nats.Conn
	rawinput     chan *nats.Msg
	tagsinput    chan *wire.Tags
	subscription *nats.Subscription
}

func New(address string, appId string) (*NatsTransport, error) {
	nc, err := nats.Connect(address)
	if err != nil {
		return nil, err
	}

	input := make(chan *nats.Msg, 16)

	subscription, err := nc.ChanSubscribe(appId, input)
	if err != nil {
		return nil, err
	}

	input2 := make(chan *wire.Tags, 16)

	// Subscription's closehandler does not work. To use connection closehandler instead.
	nc.SetClosedHandler(func(c *nats.Conn) {
		close(input)
	})

	t := NatsTransport{
		connection:   nc,
		subscription: subscription,
		rawinput:     input,
		tagsinput:    input2,
	}

	go t.listen(input, input2)

	return &t, nil
}

func (transport *NatsTransport) Send(tags wire.Tags) error {
	buf := new(bytes.Buffer)
	err := wire.Serialize(buf, tags)
	if err != nil {
		return err
	}
	header := strconv.FormatInt(int64(buf.Len()), 10) + "\n"
	buf2 := new(bytes.Buffer)
	buf2.WriteString(header)
	buf2.Write(buf.Bytes())

	_, err = transport.connection.Request(tags.Transport_Receiver, buf2.Bytes(), 10*time.Second)

	return err
}

// Listens to nats.Msg on input and forwards them as wire.Tags messages to output.
func (transport *NatsTransport) listen(input chan *nats.Msg, output chan *wire.Tags) {
	defer close(output)

	for msg := range input {
		buf := bytes.NewBuffer(msg.Data)
		lengthsString, err := buf.ReadString('\n')
		if err != nil {
			panic(err)
		}
		lengths := strings.Split(lengthsString, ",")
		for range lengths {
			tags := wire.Tags{}
			err := wire.Deserialize(buf, &tags)
			if err != nil {
				panic(err)
			}
			output <- &tags
		}
	}
}

func (transport *NatsTransport) Stop() {
	err := transport.subscription.Drain()
	if err != nil {
		panic(err)
	}
	err = transport.connection.Drain()
	if err != nil {
		panic(err)
	}
}

// Returns the channel to which incoming messages are emitted.
func (transport NatsTransport) GetInputChannel() chan *wire.Tags {
	return transport.tagsinput
}
