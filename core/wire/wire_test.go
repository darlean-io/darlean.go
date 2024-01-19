package wire

import (
	"bytes"
	"testing"

	"github.com/darlean-io/darlean.go/utils/binary"
	"github.com/darlean-io/darlean.go/utils/checks"
	"github.com/darlean-io/darlean.go/utils/variant"
)

type someStruct struct {
	AString string
	AInt    int
	AFloat  float32
	ABool   bool
}

func TestActorContainer(t *testing.T) {
	tags := TagsOut{
		TransportTags: TransportTags{
			Transport_Receiver: "Receiver",
			Transport_Return:   "Return",
		},
		RemoteCallTags: RemoteCallTags{
			Remotecall_Kind: "call",
			Remotecall_Id:   "12345",
		},
		ActorCallRequestOut: ActorCallRequestOut{
			Lazy:       true,
			ActorType:  "Type",
			ActionName: "Action",
			ActorId:    []string{"a", "b"},
			Arguments:  []any{"a", 1, true, map[string]int{"five": 5, "nine": 9}, binary.FromBytes([]byte("HELLO"))},
		},
		ActorCallResponseOut: ActorCallResponseOut{
			Error: "Error",
			Value: "Value",
		},
	}

	var buf bytes.Buffer
	Serialize(&buf, tags)
	var tags2 TagsIn
	Deserialize(&buf, &tags2)

	checks.Equal(t, "Receiver", tags2.Transport_Receiver, "Transport Receiver")
	checks.Equal(t, "Return", tags2.Transport_Return, "Transport Return")
	checks.Equal(t, "call", tags2.Remotecall_Kind, "Remotecall Kind")
	checks.Equal(t, "12345", tags2.Remotecall_Id, "Remotecall Id")
	checks.Equal(t, true, tags2.Lazy, "ActorCallRequest Lazy")
	checks.Equal(t, "Type", tags2.ActorType, "ActorCallRequest Actortype")
	checks.Equal(t, "Action", tags2.ActionName, "ActorCallRequest Actionname")
	checks.Equal(t, []string{"a", "b"}, tags2.ActorId, "ActorCallRequest Actor Id")

	checks.Equal(t, "a", tags2.Arguments[0], "ActorCallRequest Argument 0")

	var arg1 int
	tags2.Arguments[1].(variant.Assignable).AssignTo(&arg1)
	checks.Equal(t, 1, arg1, "ActorCallRequest Argument 1")

	checks.Equal(t, true, tags2.Arguments[2], "ActorCallRequest Argument 2")

	var arg3 map[string]int
	tags2.Arguments[3].(variant.Assignable).AssignTo(&arg3)
	checks.Equal(t, map[string]int{"five": 5, "nine": 9}, arg3, "ActorCallRequest Argument 3")

	checks.Equal(t, []byte("HELLO"), tags2.Arguments[4], "ActorCallRequest Argument 4")

	var error string
	tags2.Error.(variant.Assignable).AssignTo(&error)
	checks.Equal(t, "Error", error, "ActorCallResponse Error")

	checks.Equal(t, "Value", tags2.Value, "ActorCallResponse Value")
}

func TestActorContainerWithStructs(t *testing.T) {
	aStruct := someStruct{AString: "Foo", AInt: 42, AFloat: 3.1, ABool: true}
	tags := TagsOut{
		ActorCallRequestOut: ActorCallRequestOut{
			Arguments: []any{aStruct},
		},
		ActorCallResponseOut: ActorCallResponseOut{
			Error: aStruct,
			Value: aStruct,
		},
	}

	var buf bytes.Buffer
	Serialize(&buf, tags)
	var tags2 TagsIn
	Deserialize(&buf, &tags2)

	checks.Equal(t, "", tags2.Transport_Receiver, "Transport Receiver")
	checks.Equal(t, "", tags2.Transport_Return, "Transport Return")
	checks.Equal(t, "", tags2.Remotecall_Id, "Remotecall Id")
	checks.Equal(t, false, tags2.Lazy, "ActorCallRequest Lazy")
	checks.Equal(t, "", tags2.ActorType, "ActorCallRequest Actortype")
	checks.Equal(t, "", tags2.ActionName, "ActorCallRequest Actionname")
	checks.Equal(t, nil, tags2.ActorId, "ActorCallRequest Actor Id")

	var st someStruct
	tags2.Arguments[0].(variant.Assignable).AssignTo(&st)
	checks.Equal(t, "Foo", st.AString, "ActorCallRequest Argument 1")
	checks.Equal(t, 42, st.AInt, "ActorCallRequest Argument 1")
	checks.Equal(t, 3.1, st.AFloat, "ActorCallRequest Argument 1")
	checks.Equal(t, true, st.ABool, "ActorCallRequest Argument 1")

	var st1 someStruct
	tags2.Error.(variant.Assignable).AssignTo(&st1)
	checks.Equal(t, "Foo", st1.AString, "ActorCallResponse Error")
	checks.Equal(t, 42, st1.AInt, "ActorCallResponse Error")
	checks.Equal(t, 3.1, st1.AFloat, "ActorCallResponse Error")
	checks.Equal(t, true, st1.ABool, "ActorCallResponse Error")

	var st2 someStruct
	tags2.Value.(variant.Assignable).AssignTo(&st2)
	checks.Equal(t, "Foo", st2.AString, "ActorCallResponse Value")
	checks.Equal(t, 42, st2.AInt, "ActorCallResponse Value")
	checks.Equal(t, 3.1, st2.AFloat, "ActorCallResponse Value")
	checks.Equal(t, true, st2.ABool, "ActorCallResponse Value")
}
